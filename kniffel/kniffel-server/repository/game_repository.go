package repository

import (
	"github.com/oglimmer/kniffel/model"
	"gorm.io/gorm"
)

// GameRepository handles database operations for games
type GameRepository struct {
	db *gorm.DB
}

// NewGameRepository creates a new GameRepository
func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

// Save persists a KniffelGame to the database
func (r *GameRepository) Save(game *model.KniffelGame) error {
	entity := model.ToGameEntity(game)

	// check if already exists
	var existing model.GameEntity
	result := r.db.Where("game_id = ?", game.GameID).First(&existing)
	if result.Error == nil {
		// update existing: delete old related records and replace
		r.db.Where("game_entity_id = ?", existing.ID).Delete(&model.DiceRollEntity{})

		// delete used booking types for all players
		var playerIDs []uint
		r.db.Model(&model.PlayerEntity{}).Where("game_entity_id = ?", existing.ID).Pluck("id", &playerIDs)
		if len(playerIDs) > 0 {
			r.db.Where("player_entity_id IN ?", playerIDs).Delete(&model.UsedBookingTypeEntity{})
		}
		r.db.Where("game_entity_id = ?", existing.ID).Delete(&model.PlayerEntity{})

		entity.ID = existing.ID
		entity.CreatedAt = existing.CreatedAt
		return r.db.Save(entity).Error
	}

	return r.db.Create(entity).Error
}

// FindByGameID finds a game by its string game ID
func (r *GameRepository) FindByGameID(gameID string) (*model.KniffelGame, error) {
	var entity model.GameEntity
	result := r.db.
		Preload("Players", func(db *gorm.DB) *gorm.DB {
			return db.Order("player_index ASC")
		}).
		Preload("Players.UsedBookingTypes").
		Preload("DiceRolls", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Where("game_id = ?", gameID).
		First(&entity)
	if result.Error != nil {
		return nil, result.Error
	}
	return entity.ToKniffelGame(), nil
}

// DB returns the underlying database connection for health checks
func (r *GameRepository) DB() *gorm.DB {
	return r.db
}
