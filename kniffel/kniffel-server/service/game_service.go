package service

import (
	"log/slog"
	"sync"

	"github.com/oglimmer/kniffel/model"
	"github.com/oglimmer/kniffel/repository"
)

// GameService manages game lifecycle and business logic
type GameService struct {
	repo *repository.GameRepository
	// in-memory fallback when no database is configured
	mu    sync.RWMutex
	games map[string]*model.KniffelGame
}

// NewGameService creates a GameService with in-memory storage
func NewGameService() *GameService {
	return &GameService{
		games: make(map[string]*model.KniffelGame),
	}
}

// NewGameServiceWithRepo creates a GameService backed by a database repository
func NewGameServiceWithRepo(repo *repository.GameRepository) *GameService {
	return &GameService{repo: repo}
}

// CreateGame creates a new game with the given player names
func (s *GameService) CreateGame(playerNames []string) (*model.KniffelGame, error) {
	if len(playerNames) < 2 {
		return nil, &model.ErrBadRequest{Message: "at least 2 players have to be provided"}
	}

	seen := make(map[string]bool)
	for _, name := range playerNames {
		if seen[name] {
			return nil, &model.ErrBadRequest{Message: "all player names must be different"}
		}
		seen[name] = true
	}

	players := make([]model.KniffelPlayer, len(playerNames))
	for i, name := range playerNames {
		players[i] = model.KniffelPlayer{Name: name}
	}
	game := model.NewKniffelGame(players)

	slog.Info("game created", "gameID", game.GameID, "players", playerNames)

	if s.repo != nil {
		if err := s.repo.Save(game); err != nil {
			return nil, err
		}
	} else {
		s.mu.Lock()
		s.games[game.GameID] = game
		s.mu.Unlock()
	}

	return game, nil
}

// GetGameInfo retrieves a game by its ID
func (s *GameService) GetGameInfo(gameID string) (*model.KniffelGame, error) {
	if s.repo != nil {
		game, err := s.repo.FindByGameID(gameID)
		if err != nil {
			return nil, &model.ErrNotFound{Message: gameID}
		}
		return game, nil
	}

	s.mu.RLock()
	game, ok := s.games[gameID]
	s.mu.RUnlock()
	if !ok {
		return nil, &model.ErrNotFound{Message: gameID}
	}
	return game, nil
}

// Roll re-rolls dice for the current player
func (s *GameService) Roll(game *model.KniffelGame, diceToKeep []int) error {
	if game.State != model.StateRoll {
		return &model.ErrBadRequest{Message: "game is not in roll state"}
	}

	slog.Debug("rolling dice", "gameID", game.GameID, "currentDice", game.DiceRolls, "toKeep", diceToKeep)

	game.ReRollDice(diceToKeep)

	if s.repo != nil {
		return s.repo.Save(game)
	}
	return nil
}

// BookRoll books the current dice roll into a scoring category
func (s *GameService) BookRoll(game *model.KniffelGame, bookingType string) error {
	if game.State != model.StateBook {
		return &model.ErrBadRequest{Message: "game is not in book state"}
	}

	slog.Info("booking roll", "gameID", game.GameID, "bookingType", bookingType, "player", game.CurrentPlayer().Name)

	game.BookDiceRoll(bookingType)

	if s.repo != nil {
		return s.repo.Save(game)
	}
	return nil
}
