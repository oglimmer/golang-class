package model

import "gorm.io/gorm"

// GameEntity is the database entity for a Kniffel game
type GameEntity struct {
	gorm.Model
	GameID           string              `gorm:"uniqueIndex;not null"`
	CurrentPlayerIdx int
	State            string              `gorm:"not null"`
	RollRound        int
	Players          []PlayerEntity      `gorm:"foreignKey:GameEntityID"`
	DiceRolls        []DiceRollEntity    `gorm:"foreignKey:GameEntityID"`
}

func (GameEntity) TableName() string {
	return "kniffel_games"
}

// PlayerEntity is the database entity for a player
type PlayerEntity struct {
	gorm.Model
	GameEntityID     uint
	Name             string
	Score            int
	PlayerIndex      int
	UsedBookingTypes []UsedBookingTypeEntity `gorm:"foreignKey:PlayerEntityID"`
}

func (PlayerEntity) TableName() string {
	return "kniffel_players"
}

// DiceRollEntity stores individual dice values
type DiceRollEntity struct {
	gorm.Model
	GameEntityID uint
	Position     int
	Value        int
}

func (DiceRollEntity) TableName() string {
	return "kniffel_dice_rolls"
}

// UsedBookingTypeEntity stores used booking types per player
type UsedBookingTypeEntity struct {
	gorm.Model
	PlayerEntityID uint
	BookingType    string
}

func (UsedBookingTypeEntity) TableName() string {
	return "kniffel_used_booking_types"
}

// ToKniffelGame converts a GameEntity to a KniffelGame (domain model)
func (e *GameEntity) ToKniffelGame() *KniffelGame {
	players := make([]KniffelPlayer, len(e.Players))
	for _, p := range e.Players {
		usedTypes := make([]BookingType, len(p.UsedBookingTypes))
		for j, bt := range p.UsedBookingTypes {
			usedTypes[j] = BookingType(bt.BookingType)
		}
		players[p.PlayerIndex] = KniffelPlayer{
			Name:             p.Name,
			Score:            p.Score,
			UsedBookingTypes: usedTypes,
		}
	}

	diceRolls := make([]int, 5)
	for _, d := range e.DiceRolls {
		if d.Position >= 0 && d.Position < 5 {
			diceRolls[d.Position] = d.Value
		}
	}

	return &KniffelGame{
		GameID:           e.GameID,
		Players:          players,
		CurrentPlayerIdx: e.CurrentPlayerIdx,
		State:            GameState(e.State),
		DiceRolls:        diceRolls,
		RollRound:        e.RollRound,
	}
}

// ToGameEntity converts a KniffelGame (domain model) to a GameEntity
func ToGameEntity(game *KniffelGame) *GameEntity {
	players := make([]PlayerEntity, len(game.Players))
	for i, p := range game.Players {
		usedTypes := make([]UsedBookingTypeEntity, len(p.UsedBookingTypes))
		for j, bt := range p.UsedBookingTypes {
			usedTypes[j] = UsedBookingTypeEntity{BookingType: string(bt)}
		}
		players[i] = PlayerEntity{
			Name:             p.Name,
			Score:            p.Score,
			PlayerIndex:      i,
			UsedBookingTypes: usedTypes,
		}
	}

	diceRolls := make([]DiceRollEntity, len(game.DiceRolls))
	for i, v := range game.DiceRolls {
		diceRolls[i] = DiceRollEntity{Position: i, Value: v}
	}

	return &GameEntity{
		GameID:           game.GameID,
		CurrentPlayerIdx: game.CurrentPlayerIdx,
		State:            string(game.State),
		RollRound:        game.RollRound,
		Players:          players,
		DiceRolls:        diceRolls,
	}
}
