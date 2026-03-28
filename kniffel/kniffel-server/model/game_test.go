package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		dice     []int
		bt       BookingType
		expected int
	}{
		{"ones - two ones", []int{1, 1, 2, 3, 4}, Ones, 2},
		{"ones - no ones", []int{2, 3, 4, 5, 6}, Ones, 0},
		{"twos", []int{2, 2, 2, 3, 4}, Twos, 6},
		{"threes", []int{3, 3, 3, 1, 2}, Threes, 9},
		{"fours", []int{4, 4, 1, 2, 3}, Fours, 8},
		{"fives", []int{5, 5, 5, 1, 2}, Fives, 15},
		{"sixes", []int{6, 6, 6, 6, 1}, Sixes, 24},
		{"three of a kind - success", []int{1, 3, 6, 6, 6}, ThreeOfAKind, 22},
		{"three of a kind - with 4", []int{1, 1, 1, 1, 2}, ThreeOfAKind, 6},
		{"three of a kind - no match", []int{1, 3, 5, 6, 6}, ThreeOfAKind, 0},
		{"four of a kind - success", []int{3, 3, 3, 3, 1}, FourOfAKind, 13},
		{"four of a kind - no match", []int{3, 3, 3, 1, 2}, FourOfAKind, 0},
		{"full house - success", []int{2, 2, 3, 3, 3}, FullHouse, 25},
		{"full house - no match", []int{2, 2, 3, 4, 3}, FullHouse, 0},
		{"small straight 1-4", []int{1, 2, 3, 4, 6}, SmallStraight, 30},
		{"small straight 2-5", []int{2, 3, 4, 5, 1}, SmallStraight, 30},
		{"small straight 3-6", []int{3, 4, 5, 6, 1}, SmallStraight, 30},
		{"small straight - no match", []int{1, 2, 4, 5, 6}, SmallStraight, 0},
		{"large straight 1-5", []int{1, 2, 3, 4, 5}, LargeStraight, 40},
		{"large straight 2-6", []int{2, 3, 4, 5, 6}, LargeStraight, 40},
		{"large straight - no match", []int{1, 2, 3, 4, 6}, LargeStraight, 0},
		{"kniffel - success", []int{5, 5, 5, 5, 5}, Kniffel, 50},
		{"kniffel - no match", []int{5, 5, 5, 5, 1}, Kniffel, 0},
		{"chance", []int{1, 2, 3, 4, 5}, Chance, 15},
		{"chance - all sixes", []int{6, 6, 6, 6, 6}, Chance, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateScore(tt.dice, tt.bt)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestNewKniffelGame(t *testing.T) {
	players := []KniffelPlayer{
		{Name: "oli"},
		{Name: "mike"},
	}
	game := NewKniffelGame(players)

	assert.NotEmpty(t, game.GameID)
	assert.Equal(t, 2, len(game.Players))
	assert.Equal(t, "oli", game.CurrentPlayer().Name)
	assert.Equal(t, StateRoll, game.State)
	assert.Equal(t, 1, game.RollRound)
	assert.Equal(t, 5, len(game.DiceRolls))

	for _, d := range game.DiceRolls {
		assert.GreaterOrEqual(t, d, 1)
		assert.LessOrEqual(t, d, 6)
	}
}

func TestReRollDice(t *testing.T) {
	players := []KniffelPlayer{{Name: "a"}, {Name: "b"}}
	game := NewKniffelGame(players)

	game.ReRollDice([]int{})
	assert.Equal(t, 2, game.RollRound)
	assert.Equal(t, StateRoll, game.State)
	assert.Equal(t, 5, len(game.DiceRolls))

	game.ReRollDice([]int{})
	assert.Equal(t, 3, game.RollRound)
	assert.Equal(t, StateBook, game.State)
}

func TestBookDiceRoll(t *testing.T) {
	players := []KniffelPlayer{{Name: "oli"}, {Name: "mike"}}
	game := NewKniffelGame(players)

	// advance to BOOK state
	game.ReRollDice([]int{})
	game.ReRollDice([]int{})
	assert.Equal(t, StateBook, game.State)

	game.BookDiceRoll("CHANCE")
	assert.Equal(t, "mike", game.CurrentPlayer().Name)
	assert.Equal(t, StateRoll, game.State)
	assert.Equal(t, 1, game.RollRound)
	assert.Contains(t, game.Players[0].UsedBookingTypes, Chance)
}

func TestAvailableBookingTypes(t *testing.T) {
	players := []KniffelPlayer{{Name: "a"}, {Name: "b"}}
	game := NewKniffelGame(players)

	available := game.AvailableBookingTypes()
	assert.Equal(t, 13, len(available))

	// book one type
	game.ReRollDice([]int{})
	game.ReRollDice([]int{})
	game.BookDiceRoll("CHANCE")

	// back to player "a" after player "b" books
	game.ReRollDice([]int{})
	game.ReRollDice([]int{})
	game.BookDiceRoll("ONES")

	// now player "a" should have 12 available
	available = game.AvailableBookingTypes()
	assert.Equal(t, 12, len(available))
	assert.NotContains(t, available, "CHANCE")
}
