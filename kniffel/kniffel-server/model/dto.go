package model

// GameResponse is the response for all game endpoints
type GameResponse struct {
	GameID                string       `json:"gameId"`
	PlayerData            []PlayerData `json:"playerData"`
	CurrentPlayerName     string       `json:"currentPlayerName"`
	State                 string       `json:"state"`
	UsedBookingTypes      []string     `json:"usedBookingTypes"`
	AvailableBookingTypes []string     `json:"availableBookingTypes"`
	DiceRolls             []int        `json:"diceRolls"`
	RollRound             int          `json:"rollRound"`
}

// PlayerData holds player name and score
type PlayerData struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// CreateGameRequest is the request body for creating a game
type CreateGameRequest struct {
	PlayerNames []string `json:"playerNames"`
}

// DiceRollRequest is the request body for re-rolling dice
type DiceRollRequest struct {
	DiceToKeep []int `json:"diceToKeep"`
}

// BookRollRequest is the request body for booking a roll
type BookRollRequest struct {
	BookingType string `json:"bookingType"`
}
