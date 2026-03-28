package model

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"sort"
)

// BookingType represents a Kniffel scoring category
type BookingType string

const (
	Ones          BookingType = "ONES"
	Twos          BookingType = "TWOS"
	Threes        BookingType = "THREES"
	Fours         BookingType = "FOURS"
	Fives         BookingType = "FIVES"
	Sixes         BookingType = "SIXES"
	ThreeOfAKind  BookingType = "THREE_OF_A_KIND"
	FourOfAKind   BookingType = "FOUR_OF_A_KIND"
	FullHouse     BookingType = "FULL_HOUSE"
	SmallStraight BookingType = "SMALL_STRAIGHT"
	LargeStraight BookingType = "LARGE_STRAIGHT"
	Kniffel       BookingType = "KNIFFEL"
	Chance        BookingType = "CHANCE"
)

// AllBookingTypes returns all possible booking types
var AllBookingTypes = []BookingType{
	Ones, Twos, Threes, Fours, Fives, Sixes,
	ThreeOfAKind, FourOfAKind, FullHouse,
	SmallStraight, LargeStraight, Kniffel, Chance,
}

// GameState represents the current game state
type GameState string

const (
	StateRoll GameState = "ROLL"
	StateBook GameState = "BOOK"
)

// KniffelPlayer holds player data
type KniffelPlayer struct {
	Name             string
	Score            int
	UsedBookingTypes []BookingType
}

// KniffelGame is the main game logic
type KniffelGame struct {
	GameID           string
	Players          []KniffelPlayer
	CurrentPlayerIdx int
	State            GameState
	DiceRolls        []int
	RollRound        int
}

// NewKniffelGame creates a new game with the given players
func NewKniffelGame(players []KniffelPlayer) *KniffelGame {
	game := &KniffelGame{
		GameID:           generateID(),
		Players:          players,
		CurrentPlayerIdx: 0,
		State:            StateRoll,
		RollRound:        1,
	}
	game.rollAllDice()
	return game
}

// CurrentPlayer returns the current player
func (g *KniffelGame) CurrentPlayer() *KniffelPlayer {
	return &g.Players[g.CurrentPlayerIdx]
}

// ReRollDice re-rolls dice not in the keep list
func (g *KniffelGame) ReRollDice(diceToKeep []int) {
	newDice := make([]int, 0, 5)
	for _, d := range diceToKeep {
		newDice = append(newDice, d)
	}
	for len(newDice) < 5 {
		newDice = append(newDice, rollDie())
	}
	g.DiceRolls = newDice
	g.RollRound++

	if g.RollRound >= 3 {
		g.State = StateBook
	}
}

// BookDiceRoll books the current dice into a scoring category
func (g *KniffelGame) BookDiceRoll(bookingType string) {
	bt := BookingType(bookingType)
	player := g.CurrentPlayer()

	score := CalculateScore(g.DiceRolls, bt)
	player.Score += score
	player.UsedBookingTypes = append(player.UsedBookingTypes, bt)

	g.CurrentPlayerIdx = (g.CurrentPlayerIdx + 1) % len(g.Players)
	g.State = StateRoll
	g.RollRound = 1
	g.rollAllDice()
}

// AvailableBookingTypes returns booking types not yet used by the current player
func (g *KniffelGame) AvailableBookingTypes() []string {
	player := g.CurrentPlayer()
	used := make(map[BookingType]bool)
	for _, bt := range player.UsedBookingTypes {
		used[bt] = true
	}

	available := []string{}
	for _, bt := range AllBookingTypes {
		if !used[bt] {
			available = append(available, string(bt))
		}
	}
	return available
}

// UsedBookingTypesAsStrings returns used booking types as strings
func (g *KniffelGame) UsedBookingTypesAsStrings() []string {
	player := g.CurrentPlayer()
	used := []string{}
	for _, bt := range player.UsedBookingTypes {
		used = append(used, string(bt))
	}
	return used
}

func (g *KniffelGame) rollAllDice() {
	g.DiceRolls = make([]int, 5)
	for i := range g.DiceRolls {
		g.DiceRolls[i] = rollDie()
	}
}

func rollDie() int {
	n, _ := rand.Int(rand.Reader, big.NewInt(6))
	return int(n.Int64()) + 1
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CalculateScore computes the score for a given dice roll and booking type
func CalculateScore(dice []int, bt BookingType) int {
	sum := 0
	for _, d := range dice {
		sum += d
	}

	switch bt {
	case Ones:
		return countValue(dice, 1)
	case Twos:
		return countValue(dice, 2) * 2
	case Threes:
		return countValue(dice, 3) * 3
	case Fours:
		return countValue(dice, 4) * 4
	case Fives:
		return countValue(dice, 5) * 5
	case Sixes:
		return countValue(dice, 6) * 6
	case ThreeOfAKind:
		if hasNOfAKind(dice, 3) {
			return sum
		}
		return 0
	case FourOfAKind:
		if hasNOfAKind(dice, 4) {
			return sum
		}
		return 0
	case FullHouse:
		if isFullHouse(dice) {
			return 25
		}
		return 0
	case SmallStraight:
		if isSmallStraight(dice) {
			return 30
		}
		return 0
	case LargeStraight:
		if isLargeStraight(dice) {
			return 40
		}
		return 0
	case Kniffel:
		if hasNOfAKind(dice, 5) {
			return 50
		}
		return 0
	case Chance:
		return sum
	}
	return 0
}

func countValue(dice []int, value int) int {
	count := 0
	for _, d := range dice {
		if d == value {
			count++
		}
	}
	return count
}

func hasNOfAKind(dice []int, n int) bool {
	counts := map[int]int{}
	for _, d := range dice {
		counts[d]++
	}
	for _, c := range counts {
		if c >= n {
			return true
		}
	}
	return false
}

func isFullHouse(dice []int) bool {
	counts := map[int]int{}
	for _, d := range dice {
		counts[d]++
	}
	hasTwo, hasThree := false, false
	for _, c := range counts {
		if c == 2 {
			hasTwo = true
		}
		if c == 3 {
			hasThree = true
		}
	}
	return hasTwo && hasThree
}

func isSmallStraight(dice []int) bool {
	unique := uniqueSorted(dice)
	straights := [][]int{{1, 2, 3, 4}, {2, 3, 4, 5}, {3, 4, 5, 6}}
	for _, s := range straights {
		if containsAll(unique, s) {
			return true
		}
	}
	return false
}

func isLargeStraight(dice []int) bool {
	unique := uniqueSorted(dice)
	straights := [][]int{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}}
	for _, s := range straights {
		if containsAll(unique, s) {
			return true
		}
	}
	return false
}

func uniqueSorted(dice []int) []int {
	set := map[int]bool{}
	for _, d := range dice {
		set[d] = true
	}
	result := make([]int, 0, len(set))
	for v := range set {
		result = append(result, v)
	}
	sort.Ints(result)
	return result
}

func containsAll(haystack, needles []int) bool {
	set := map[int]bool{}
	for _, v := range haystack {
		set[v] = true
	}
	for _, n := range needles {
		if !set[n] {
			return false
		}
	}
	return true
}
