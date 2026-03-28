# Chapter 03

Swagger, Application layers and game logic

# Goal

Adding OpenAPI documentation, understanding application layers, dependency injection in Go and implementing a GameService.

# Context and Knowledge

* For a REST API application you usually want to separate your code into at least 3 layers:
    * **presentation** layer, also called HTTP, web, REST layer (in Go: `handler/` package)
    * **business** layer, also called logic, services layer (in Go: `service/` package)
    * **persistence** layer, also called database, data access layer (in Go: `repository/` package)
* The handler functions and the DTOs used in the handlers live in the presentation layer
* The business layer contains the business logic or in our case the game logic and the services to manage them
* The persistence layer has the code to access the database, there should be no business logic there

## DB transaction control

If you don't know what DB transactions are, just skip this.

One reason to use different application layers is to have control over the DB transactions:

* presentation layer = all functions here never participate in transactions
* business layer = each function controls its transactions
* persistence layer = each function either participates in an existing transaction or starts a new one if none existed

We'll come back to this in chapter 06.

## Follow up from Chapter 02 - Review the 4 endpoints

Before we continue we want to review the "homework" from the last step in Chapter 02.

First, let's organize our project into packages. Create the following directory structure:

```
kniffel/
├── go.mod
├── go.sum
├── main.go
├── handler/
│   └── game_handler.go
├── service/
│   └── game_service.go
└── model/
    ├── dto.go
    └── game.go
```

### The DTOs (model/dto.go)

In Go, DTOs are just structs. No Lombok annotations needed - Go structs are already concise:

```go
package model

// GameResponse is the response for all game endpoints
type GameResponse struct {
    GameID                string       `json:"gameId"`
    PlayerData            []PlayerData `json:"playerData"`
    CurrentPlayerName     string       `json:"currentPlayerName"`
    State                 string       `json:"state"`     // "BOOK" or "ROLL"
    UsedBookingTypes      []string     `json:"usedBookingTypes"`
    AvailableBookingTypes []string     `json:"availableBookingTypes"`
    DiceRolls             []int        `json:"diceRolls"` // dice values (1-6), array size = 5
    RollRound             int          `json:"rollRound"`  // for state==ROLL: round 1, 2 or 3
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
```

Notice that in Go we don't need separate files for each struct. Related types can live in the same file. Also, no getters or setters are needed - Go struct fields are accessed directly.

### The Game Handler (handler/game_handler.go)

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/oglimmer/kniffel/model"
    "github.com/oglimmer/kniffel/service"
)

// GameServiceInterface defines what the handler needs from the service layer.
// In Go, interfaces are defined by the consumer, not the provider.
type GameServiceInterface interface {
    CreateGame(playerNames []string) *model.KniffelGame
    GetGameInfo(gameID string) *model.KniffelGame
    Roll(game *model.KniffelGame, diceToKeep []int)
    BookRoll(game *model.KniffelGame, bookingType string)
}

type GameHandler struct {
    gameService GameServiceInterface
}

func NewGameHandler(gameService GameServiceInterface) *GameHandler {
    return &GameHandler{gameService: gameService}
}

func (h *GameHandler) RegisterRoutes(r *gin.RouterGroup) {
    r.POST("/", h.CreateGame)
    r.GET("/:gameID", h.GetGameInfo)
    r.POST("/:gameID/roll", h.Roll)
    r.POST("/:gameID/book", h.Book)
}

func (h *GameHandler) CreateGame(c *gin.Context) {
    var req model.CreateGameRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // TODO: implement
    c.JSON(http.StatusOK, model.GameResponse{})
}

func (h *GameHandler) GetGameInfo(c *gin.Context) {
    gameID := c.Param("gameID")
    _ = gameID // TODO: implement
    c.JSON(http.StatusOK, model.GameResponse{})
}

func (h *GameHandler) Roll(c *gin.Context) {
    gameID := c.Param("gameID")
    var req model.DiceRollRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    _ = gameID // TODO: implement
    c.JSON(http.StatusOK, model.GameResponse{})
}

func (h *GameHandler) Book(c *gin.Context) {
    gameID := c.Param("gameID")
    var req model.BookRollRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    _ = gameID // TODO: implement
    c.JSON(http.StatusOK, model.GameResponse{})
}
```

If your structs and handlers look a bit different, you can either keep them as they are or change them to my suggestion. Using my API definition gives you the chance to use my Web UI later on.

## Step 1 - Swagger / OpenAPI

REST APIs are often provided by team A and used by team B - even outside the company, so good documentation is very important.

There is a standard for documenting REST APIs: [OpenAPI](https://spec.openapis.org/oas/latest.html). A company called SmartBear provides a software called Swagger, a web UI for OpenAPI documentation.

In Go, we use `swaggo/swag` to generate OpenAPI docs from code comments.

### Install swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`.

### Add dependencies

```bash
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

### Add Swagger annotations

In swaggo, you document your API using special comments. First, add general API info to `main.go`:

```go
// @title Kniffel Game API
// @version 0.0.1
// @description Kniffel as a service - KaaS
// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0

// @externalDocs.description Kniffel Regeln
// @externalDocs.url https://www.schmidtspiele.de/files/Produkte/4/49030%20-%20Kniffel/49203_49030_Kniffel_DE.pdf
func main() {
    // ...
}
```

Then annotate each handler function:

```go
// CreateGame creates a new game
// @Summary Create a new game with a specific number of players
// @Description The number of players must be at least 2. All player names must be different.
// @Tags game
// @Accept json
// @Produce json
// @Param request body model.CreateGameRequest true "Game creation request"
// @Success 200 {object} model.GameResponse
// @Router /api/v1/game/ [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
    // ... handler code
}
```

### Generate and serve the docs

Run the swagger generation:

```bash
swag init
```

This creates a `docs/` directory with the generated OpenAPI spec.

Now register the Swagger UI in `main.go`:

```go
import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "github.com/oglimmer/kniffel/docs" // import generated docs
)

func main() {
    r := gin.Default()
    r.Use(cors.Default())

    // Swagger UI
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // ... routes
}
```

Access http://localhost:8080/swagger/index.html and you should see your API documentation.

The raw OpenAPI JSON is available at http://localhost:8080/swagger/doc.json.

**Important**: Run `swag init` every time you change the Swagger annotations!

## Step 2 - Dependency Injection in Go

Go does not have a built-in DI framework like Spring's `@Autowired`. Instead, Go uses **constructor injection** - you pass dependencies as function parameters. This is simpler and more explicit.

In Spring you would write:

```java
@Service
public class GameService {
    @Autowired
    private GameRepository gameRepository;
}
```

In Go, the equivalent pattern is:

```go
type GameService struct {
    // dependency stored as a field
    games map[string]*KniffelGame
}

// NewGameService is the "constructor" - this is where DI happens
func NewGameService() *GameService {
    return &GameService{
        games: make(map[string]*KniffelGame),
    }
}
```

And in `main.go`, you wire everything together manually:

```go
func main() {
    r := gin.Default()
    r.Use(cors.Default())

    // "Dependency Injection" - we create instances and pass them explicitly
    gameService := service.NewGameService()
    gameHandler := handler.NewGameHandler(gameService)

    // register routes
    gameGroup := r.Group("/api/v1/game")
    gameHandler.RegisterRoutes(gameGroup)

    r.Run(":8080")
}
```

This approach is sometimes called "poor man's DI" but it's the idiomatic Go way. It has benefits:

* **Explicit**: You can see all dependencies in the constructor
* **Testable**: Easy to pass mock implementations
* **No magic**: No reflection, no annotations, no framework

### Interfaces for DI

When you want to swap implementations (e.g., for testing), Go uses **interfaces**. An important Go principle is: **define interfaces at the consumer, not the provider** ("accept interfaces, return structs"). The interface should live in the package that _uses_ it, not in the package that _implements_ it. This keeps packages decoupled:

```go
// In the service package: define the interface the service needs
// This interface is defined here because the service is the consumer of the repository
type GameRepository interface {
    Save(game *KniffelGame) error
    FindByGameID(gameID string) (*KniffelGame, error)
}

// GameService depends on the interface, not a concrete type
type GameService struct {
    repo GameRepository
}

func NewGameService(repo GameRepository) *GameService {
    return &GameService{repo: repo}
}
```

In Go, interfaces are **implicit** - a type implements an interface simply by having the required methods. No `implements` keyword needed. This means the repository implementation doesn't even need to know about the interface - it just needs to have the right methods.

## Step 3 - Application Layers

Our REST API does everything in memory, so we don't have anything in the persistence layer.

It is good practice to have a business layer service for our domain objects. In Go we create a `GameService`:

```go
package service

import (
    "sync"
    "github.com/oglimmer/kniffel/model"
)

type GameService struct {
    // This map acts as our "database backend"
    // key = game ID, value = the actual game object
    mu    sync.RWMutex
    games map[string]*model.KniffelGame
}

func NewGameService() *GameService {
    return &GameService{
        games: make(map[string]*model.KniffelGame),
    }
}

func (s *GameService) CreateGame(playerNames []string) *model.KniffelGame {
    // create players
    players := make([]model.KniffelPlayer, len(playerNames))
    for i, name := range playerNames {
        players[i] = model.KniffelPlayer{Name: name}
    }

    // create game
    game := model.NewKniffelGame(players)

    // store in memory - lock needed as Go maps are not safe for concurrent access
    s.mu.Lock()
    s.games[game.GameID] = game
    s.mu.Unlock()
    return game
}

func (s *GameService) GetGameInfo(gameID string) *model.KniffelGame {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.games[gameID]
}

func (s *GameService) Roll(game *model.KniffelGame, diceToKeep []int) {
    game.ReRollDice(diceToKeep)
}

func (s *GameService) BookRoll(game *model.KniffelGame, bookingType string) {
    game.BookDiceRoll(bookingType)
}
```

**Important**: Go maps are **not safe for concurrent access**. Since Gin handles each HTTP request in its own goroutine, multiple requests can read/write the map simultaneously, causing a data race. We use `sync.RWMutex` to protect map access: `RLock`/`RUnlock` for reads, `Lock`/`Unlock` for writes.

## Step 4 - Game Logic - `KniffelGame`

As this is a tutorial for Go REST APIs and not for Go algorithms, I leave you with two options for the actual game logic:

* **[moderate]** Implement the game logic by yourself
* **[easy]** Use the reference implementation below as a starting point

### Game model (model/game.go)

Here is a basic skeleton. You need to implement the scoring logic:

```go
package model

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "math/big"
)

// BookingType represents a Kniffel scoring category
type BookingType string

const (
    Ones           BookingType = "ONES"
    Twos           BookingType = "TWOS"
    Threes         BookingType = "THREES"
    Fours          BookingType = "FOURS"
    Fives          BookingType = "FIVES"
    Sixes          BookingType = "SIXES"
    ThreeOfAKind   BookingType = "THREE_OF_A_KIND"
    FourOfAKind    BookingType = "FOUR_OF_A_KIND"
    FullHouse      BookingType = "FULL_HOUSE"
    SmallStraight  BookingType = "SMALL_STRAIGHT"
    LargeStraight  BookingType = "LARGE_STRAIGHT"
    Kniffel        BookingType = "KNIFFEL"
    Chance         BookingType = "CHANCE"
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
    Name             string        `json:"name"`
    Score            int           `json:"score"`
    UsedBookingTypes []BookingType `json:"-"`
}

// KniffelGame is the main game logic class
type KniffelGame struct {
    GameID           string
    Players          []KniffelPlayer
    CurrentPlayerIdx int
    State            GameState
    DiceRolls        []int
    RollRound        int
}

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

func (g *KniffelGame) CurrentPlayer() *KniffelPlayer {
    return &g.Players[g.CurrentPlayerIdx]
}

func (g *KniffelGame) ReRollDice(diceToKeep []int) {
    // build new dice: keep specified dice, re-roll the rest
    kept := make(map[int]int) // value -> count of kept
    for _, d := range diceToKeep {
        kept[d]++
    }

    newDice := make([]int, 0, 5)
    // first add kept dice
    for _, d := range diceToKeep {
        newDice = append(newDice, d)
    }
    // fill remaining with new random dice
    for len(newDice) < 5 {
        newDice = append(newDice, rollDie())
    }
    g.DiceRolls = newDice
    g.RollRound++

    if g.RollRound > 3 {
        g.State = StateBook
    }
}

func (g *KniffelGame) BookDiceRoll(bookingType string) {
    bt := BookingType(bookingType)
    player := g.CurrentPlayer()

    // calculate score - implement your own scoring logic here!
    score := calculateScore(g.DiceRolls, bt)
    player.Score += score
    player.UsedBookingTypes = append(player.UsedBookingTypes, bt)

    // move to next player
    g.CurrentPlayerIdx = (g.CurrentPlayerIdx + 1) % len(g.Players)
    g.State = StateRoll
    g.RollRound = 1
    g.rollAllDice()
}

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

func (g *KniffelGame) UsedBookingTypes() []string {
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
    if _, err := rand.Read(b); err != nil {
        panic(fmt.Sprintf("failed to generate random ID: %v", err))
    }
    return hex.EncodeToString(b)
}

func calculateScore(dice []int, bt BookingType) int {
    // TODO: implement proper scoring logic for each booking type
    // For now, a simplified version:
    sum := 0
    for _, d := range dice {
        sum += d
    }

    switch bt {
    case Ones:
        return countValue(dice, 1) * 1
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
        if hasNOfAKind(dice, 3) { return sum }
        return 0
    case FourOfAKind:
        if hasNOfAKind(dice, 4) { return sum }
        return 0
    case FullHouse:
        if isFullHouse(dice) { return 25 }
        return 0
    case SmallStraight:
        if isSmallStraight(dice) { return 30 }
        return 0
    case LargeStraight:
        if isLargeStraight(dice) { return 40 }
        return 0
    case Kniffel:
        if hasNOfAKind(dice, 5) { return 50 }
        return 0
    case Chance:
        return sum
    }
    return 0
}

func countValue(dice []int, value int) int {
    count := 0
    for _, d := range dice {
        if d == value { count++ }
    }
    return count
}

func hasNOfAKind(dice []int, n int) bool {
    counts := map[int]int{}
    for _, d := range dice { counts[d]++ }
    for _, c := range counts {
        if c >= n { return true }
    }
    return false
}

func isFullHouse(dice []int) bool {
    counts := map[int]int{}
    for _, d := range dice { counts[d]++ }
    hasTwo, hasThree := false, false
    for _, c := range counts {
        if c == 2 { hasTwo = true }
        if c == 3 { hasThree = true }
    }
    return hasTwo && hasThree
}

func isSmallStraight(dice []int) bool {
    set := map[int]bool{}
    for _, d := range dice { set[d] = true }
    // check for 4 consecutive values
    straights := [][]int{{1,2,3,4}, {2,3,4,5}, {3,4,5,6}}
    for _, s := range straights {
        found := true
        for _, v := range s {
            if !set[v] { found = false; break }
        }
        if found { return true }
    }
    return false
}

func isLargeStraight(dice []int) bool {
    set := map[int]bool{}
    for _, d := range dice { set[d] = true }
    straights := [][]int{{1,2,3,4,5}, {2,3,4,5,6}}
    for _, s := range straights {
        found := true
        for _, v := range s {
            if !set[v] { found = false; break }
        }
        if found { return true }
    }
    return false
}
```

## Step 5 - Mapping game data to response

In Java you used ModelMapper to copy data between classes. In Go, we just write a mapping function. Go keeps things explicit:

```go
// In handler/game_handler.go

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
        UsedBookingTypes:      game.UsedBookingTypes(),
        AvailableBookingTypes: game.AvailableBookingTypes(),
        DiceRolls:             game.DiceRolls,
        RollRound:             game.RollRound,
    }
}
```

Now update the `CreateGame` handler to use it:

```go
func (h *GameHandler) CreateGame(c *gin.Context) {
    var req model.CreateGameRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    game := h.gameService.CreateGame(req.PlayerNames)
    c.JSON(http.StatusOK, mapGameResponse(game))
}
```

At this point we have a working REST API to create a game of Kniffel.

## Step 6 - Complete the other endpoints

Complete the remaining handlers (`GetGameInfo`, `Roll`, `Book`) in the `GameHandler`. Each should:

1. Extract path variables and request body
2. Call the appropriate `GameService` method
3. Return the mapped `GameResponse`

All of these handlers can be implemented in less than 10 lines of code each.

# What we've learnt

* What OpenAPI and Swagger is and how to add it to a Go/Gin project using swaggo
* The concept of application layers in Go
* How dependency injection works in Go (constructor injection vs Spring's @Autowired)
* Go interfaces are implicit - no `implements` keyword
* How to write services to manage our business (game) logic
* In Go, mapping between types is done explicitly - no ModelMapper needed
* Go struct tags control JSON serialization

# Extras if you have time

* Read more about swaggo at https://github.com/swaggo/swag
* Read about Go interfaces: https://go.dev/tour/methods/9
* Read the official Go project layout guidance: https://go.dev/doc/modules/layout
