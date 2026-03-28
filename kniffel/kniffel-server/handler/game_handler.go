package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oglimmer/kniffel/model"
)

// GameServiceInterface defines what the handler needs from the service layer.
// In Go, interfaces are defined by the consumer, not the provider.
type GameServiceInterface interface {
	CreateGame(playerNames []string) (*model.KniffelGame, error)
	GetGameInfo(gameID string) (*model.KniffelGame, error)
	Roll(game *model.KniffelGame, diceToKeep []int) error
	BookRoll(game *model.KniffelGame, bookingType string) error
}

// GameHandler handles game-related HTTP requests
type GameHandler struct {
	gameService GameServiceInterface
}

// NewGameHandler creates a new GameHandler
func NewGameHandler(gameService GameServiceInterface) *GameHandler {
	return &GameHandler{gameService: gameService}
}

// RegisterRoutes registers all game routes on the given router group
func (h *GameHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/", h.CreateGame)
	r.GET("/:gameID", h.GetGameInfo)
	r.POST("/:gameID/roll", h.Roll)
	r.POST("/:gameID/book", h.Book)
}

// CreateGame creates a new game
// @Summary Create a new game with a specific number of players
// @Description The number of players must be at least 2. All player names must be different.
// @Tags game
// @Accept json
// @Produce json
// @Param request body model.CreateGameRequest true "Game creation request"
// @Success 201 {object} model.GameResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/game/ [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req model.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.CreateGame(req.PlayerNames)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, mapGameResponse(game))
}

// GetGameInfo retrieves game information
// @Summary Get game information
// @Description Retrieve the current state of a game
// @Tags game
// @Produce json
// @Param gameID path string true "Game ID"
// @Success 200 {object} model.GameResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/game/{gameID} [get]
func (h *GameHandler) GetGameInfo(c *gin.Context) {
	game, err := h.gameService.GetGameInfo(c.Param("gameID"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapGameResponse(game))
}

// Roll re-rolls dice for the current player
// @Summary Re-roll dice
// @Description Re-roll dice that the player doesn't want to keep
// @Tags game
// @Accept json
// @Produce json
// @Param gameID path string true "Game ID"
// @Param request body model.DiceRollRequest true "Dice to keep"
// @Success 200 {object} model.GameResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/game/{gameID}/roll [post]
func (h *GameHandler) Roll(c *gin.Context) {
	game, err := h.gameService.GetGameInfo(c.Param("gameID"))
	if err != nil {
		handleError(c, err)
		return
	}

	var req model.DiceRollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.gameService.Roll(game, req.DiceToKeep); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapGameResponse(game))
}

// Book books the current dice roll into a scoring category
// @Summary Book a dice roll
// @Description Book the current dice roll into a scoring category
// @Tags game
// @Accept json
// @Produce json
// @Param gameID path string true "Game ID"
// @Param request body model.BookRollRequest true "Booking type"
// @Success 200 {object} model.GameResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/game/{gameID}/book [post]
func (h *GameHandler) Book(c *gin.Context) {
	game, err := h.gameService.GetGameInfo(c.Param("gameID"))
	if err != nil {
		handleError(c, err)
		return
	}

	var req model.BookRollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.gameService.BookRoll(game, req.BookingType); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapGameResponse(game))
}

func mapGameResponse(game *model.KniffelGame) model.GameResponse {
	playerData := make([]model.PlayerData, len(game.Players))
	for i, p := range game.Players {
		playerData[i] = model.PlayerData{
			Name:  p.Name,
			Score: p.Score,
		}
	}

	return model.GameResponse{
		GameID:                game.GameID,
		PlayerData:            playerData,
		CurrentPlayerName:     game.CurrentPlayer().Name,
		State:                 string(game.State),
		UsedBookingTypes:      game.UsedBookingTypesAsStrings(),
		AvailableBookingTypes: game.AvailableBookingTypes(),
		DiceRolls:             game.DiceRolls,
		RollRound:             game.RollRound,
	}
}

func handleError(c *gin.Context, err error) {
	var badReq *model.ErrBadRequest
	var notFound *model.ErrNotFound

	switch {
	case errors.As(err, &badReq):
		c.JSON(http.StatusBadRequest, gin.H{"error": badReq.Message})
	case errors.As(err, &notFound):
		c.JSON(http.StatusNotFound, gin.H{"error": notFound.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
