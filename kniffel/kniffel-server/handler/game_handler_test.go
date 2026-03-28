package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/oglimmer/kniffel/model"
	"github.com/oglimmer/kniffel/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	gameService := service.NewGameService()
	gameHandler := NewGameHandler(gameService)

	gameGroup := r.Group("/api/v1/game")
	gameHandler.RegisterRoutes(gameGroup)

	return r
}

func createGame(t *testing.T, router *gin.Engine, playerNames []string) model.GameResponse {
	body := model.CreateGameRequest{PlayerNames: playerNames}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/game/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response model.GameResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	return response
}

func TestCreateGame_Success(t *testing.T) {
	router := setupRouter()
	response := createGame(t, router, []string{"oli", "ilo"})

	assert.NotEmpty(t, response.GameID)
	assert.Equal(t, "oli", response.CurrentPlayerName)
	assert.Equal(t, 2, len(response.PlayerData))
	assert.Equal(t, "ROLL", response.State)
	assert.Equal(t, 1, response.RollRound)
	assert.Equal(t, 5, len(response.DiceRolls))
	assert.Equal(t, 13, len(response.AvailableBookingTypes))
}

func TestCreateGame_FailDuplicate(t *testing.T) {
	router := setupRouter()

	body := model.CreateGameRequest{PlayerNames: []string{"oli", "oli"}}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/game/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGame_FailOnly1Name(t *testing.T) {
	router := setupRouter()

	body := model.CreateGameRequest{PlayerNames: []string{"oli"}}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/game/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGameInfo_Success(t *testing.T) {
	router := setupRouter()
	game := createGame(t, router, []string{"oli", "mike"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/game/"+game.GameID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.GameResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, game.GameID, response.GameID)
}

func TestGetGameInfo_NotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/game/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRoll_Success(t *testing.T) {
	router := setupRouter()
	game := createGame(t, router, []string{"oli", "mike"})

	body := model.DiceRollRequest{DiceToKeep: []int{}}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/game/"+game.GameID+"/roll", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.GameResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.RollRound)
	assert.Equal(t, "ROLL", response.State)
}

func TestFullGameTurn(t *testing.T) {
	router := setupRouter()
	game := createGame(t, router, []string{"oli", "mike"})

	// Roll 1
	rollBody, _ := json.Marshal(model.DiceRollRequest{DiceToKeep: []int{}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/game/"+game.GameID+"/roll", bytes.NewBuffer(rollBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Roll 2 - should transition to BOOK
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/game/"+game.GameID+"/roll", bytes.NewBuffer(rollBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var rollResp model.GameResponse
	json.Unmarshal(w.Body.Bytes(), &rollResp)
	assert.Equal(t, "BOOK", rollResp.State)

	// Book
	bookBody, _ := json.Marshal(model.BookRollRequest{BookingType: "CHANCE"})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/game/"+game.GameID+"/book", bytes.NewBuffer(bookBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var bookResp model.GameResponse
	json.Unmarshal(w.Body.Bytes(), &bookResp)
	assert.Equal(t, "mike", bookResp.CurrentPlayerName)
	assert.Equal(t, "ROLL", bookResp.State)
	assert.Equal(t, 1, bookResp.RollRound)
	assert.Greater(t, bookResp.PlayerData[0].Score, 0) // oli should have scored something
}

func TestBook_WrongState(t *testing.T) {
	router := setupRouter()
	game := createGame(t, router, []string{"oli", "mike"})

	// Try to book when state is ROLL (should fail)
	bookBody, _ := json.Marshal(model.BookRollRequest{BookingType: "CHANCE"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/game/"+game.GameID+"/book", bytes.NewBuffer(bookBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
