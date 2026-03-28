# Chapter 05

Testing, HTTP status codes and defensive code

# Goal

We want to use appropriate HTTP status codes and validate input parameters properly. This chapter shows you Go's approach to error handling and gives an introduction into unit and integration testing.

# Context and Knowledge

* REST APIs should use the HTTP status codes where possible. Here is a list of selected, useful return codes:
    * 200, OK -> the operation succeeded (Gin returns 200 by default with `c.JSON`)
    * 201, CREATED -> the operation (usually a POST) created something
    * 204, NO CONTENT -> while the operation succeeded, there is no content in the response
    * 400, BAD REQUEST -> the parameters provided were wrong, missing or cannot be decoded
    * 401, UNAUTHORIZED -> this method needs authentication and this failed
    * 403, FORBIDDEN -> the user is authenticated, but not authorized
    * 404, NOT FOUND -> the resource / data behind this endpoint does not exist
    * 500, INTERNAL SERVER ERROR -> something went wrong during processing
* Go does NOT have exceptions. Instead, Go uses **explicit error returns**. Functions that can fail return an `error` as their last return value. This is fundamentally different from Java's try/catch approach and makes error handling very visible.
* Automated testing is a very important part of professional software development. Go has excellent built-in testing support - no external framework needed for the basics.

# Step 1 - HTTP status codes

An endpoint which creates an entity should return 201 not 200. Let's change the status code for `CreateGame`:

```go
func (h *GameHandler) CreateGame(c *gin.Context) {
    var req model.CreateGameRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    game := h.gameService.CreateGame(req.PlayerNames)
    // use http.StatusCreated (201) instead of http.StatusOK (200)
    c.JSON(http.StatusCreated, mapGameResponse(game))
}
```

You can see the changed return code by using:

```bash
# the -v is important to see the response headers!
curl "http://localhost:8080/api/v1/game/" -d '{"playerNames": ["oli","mike"]}' -H "Content-Type: application/json" -v
```

## Validations and error handling

The user needs to provide at least 2 names and they must all be different. Our API should check that and return appropriate HTTP status codes.

### Error handling in Go - the basics

Go doesn't have exceptions. Instead, functions return errors:

```go
// Go style - errors are return values
func doSomething() error {
    if somethingWrong {
        return errors.New("something went wrong")
    }
    return nil // nil means no error
}

// caller must check the error
if err := doSomething(); err != nil {
    // handle error
}
```

### Custom error types

Let's create custom error types for our application. Create a file `model/errors.go`:

```go
package model

import "fmt"

// ErrBadRequest represents a 400 error
type ErrBadRequest struct {
    Message string
}

func (e *ErrBadRequest) Error() string {
    return e.Message
}

// ErrNotFound represents a 404 error
type ErrNotFound struct {
    Message string
}

func (e *ErrNotFound) Error() string {
    return fmt.Sprintf("not found: %s", e.Message)
}
```

In Go, any type that implements the `Error() string` method satisfies the `error` interface. No `extends Exception` needed.

### Adding validation to the service

Update `GameService` to return errors:

```go
func (s *GameService) CreateGame(playerNames []string) (*model.KniffelGame, error) {
    if len(playerNames) < 2 {
        return nil, &model.ErrBadRequest{Message: "at least 2 players have to be provided"}
    }

    // check for duplicate names
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
    s.games[game.GameID] = game
    return game, nil
}

func (s *GameService) GetGameInfo(gameID string) (*model.KniffelGame, error) {
    s.mu.RLock()
    game, ok := s.games[gameID]
    s.mu.RUnlock()
    if !ok {
        return nil, &model.ErrNotFound{Message: gameID}
    }
    return game, nil
}
```

### Handling errors in the handler

Now update the handler to check errors and return appropriate HTTP status codes:

```go
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
```

### Centralized error handling

Create a helper function that maps error types to HTTP status codes:

```go
func handleError(c *gin.Context, err error) {
    var badReq *model.ErrBadRequest
    var notFound *model.ErrNotFound

    switch {
    case errors.As(err, &badReq):
        c.JSON(http.StatusBadRequest, gin.H{"error": badReq.Message})
    case errors.As(err, &notFound):
        c.JSON(http.StatusNotFound, gin.H{"error": notFound.Message})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

`errors.As` is Go's way to check if an error is of a specific type - similar to `instanceof` in Java or catching a specific exception type.

You can test this with curl:

```bash
# the -v is important to see the response headers!
curl "http://localhost:8080/api/v1/game/" -d '{"playerNames": ["oli","oli"]}' -H "Content-Type: application/json" -v
```

# Step 2 - Error handling from a different application layer

Our business / game logic could return errors as well. Let's make sure that re-rolling can only be done when the game is in state ROLL and booking can only be done while in state BOOK.

Update the `GameService`:

```go
func (s *GameService) Roll(game *model.KniffelGame, diceToKeep []int) error {
    if game.State != model.StateRoll {
        return &model.ErrBadRequest{Message: "game is not in roll state"}
    }
    game.ReRollDice(diceToKeep)
    return nil
}

func (s *GameService) BookRoll(game *model.KniffelGame, bookingType string) error {
    if game.State != model.StateBook {
        return &model.ErrBadRequest{Message: "game is not in book state"}
    }
    game.BookDiceRoll(bookingType)
    return nil
}
```

And update the handlers accordingly to check and handle these errors using `handleError`.

At this point it is a good exercise to update all handlers to properly handle errors from the service layer.

# Step 3 - `nil` checks

Go has a similar problem to Java's `NullPointerException`: the **nil pointer dereference**. If you call a method on a nil pointer, your program panics (crashes).

However, Go addresses this differently than Java:

### The "comma, ok" pattern

Go's idiomatic way to handle potentially missing values:

```go
// map lookup - ok is false if the key doesn't exist
game, ok := s.games[gameId]
if !ok {
    return nil, &model.ErrNotFound{Message: gameId}
}
```

### Pointer checks

Always check if a pointer could be nil before using it:

```go
func (s *GameService) Roll(game *model.KniffelGame, diceToKeep []int) error {
    if game == nil {
        return &model.ErrNotFound{Message: "game is nil"}
    }
    // safe to use game now
    game.ReRollDice(diceToKeep)
    return nil
}
```

### Go's approach vs Java's

| Java | Go |
|------|-----|
| `NullPointerException` (runtime) | nil pointer dereference (panic) |
| `@NonNull` annotation (compile hint) | Not needed - check explicitly |
| `Optional<T>` | Return `(T, error)` or `(T, bool)` |
| `try/catch` | `if err != nil` |

Go's approach is more verbose but makes null/nil handling very visible. You always know where a nil check happens.

# Step 4 - Unit tests

> Unit tests check the correctness of a single function, mostly algorithms or calculations

Go has built-in testing support. Test files must end with `_test.go` and live next to the code they test.

Create a file `model/scoring_test.go`:

```go
package model

import (
    "testing"
)

func TestCountValue(t *testing.T) {
    dice := []int{1, 3, 6, 6, 6}
    result := countValue(dice, 6)
    if result != 3 {
        t.Errorf("expected 3, got %d", result)
    }
}

func TestHasThreeOfAKind_Success(t *testing.T) {
    dice := []int{1, 3, 6, 6, 6}
    if !hasNOfAKind(dice, 3) {
        t.Error("expected three of a kind to be true")
    }
}

func TestHasThreeOfAKind_Fail(t *testing.T) {
    dice := []int{1, 3, 5, 6, 6}
    if hasNOfAKind(dice, 3) {
        t.Error("expected three of a kind to be false")
    }
}

func TestCalculateScore_ThreeOfAKind(t *testing.T) {
    dice := []int{1, 3, 6, 6, 6}
    score := calculateScore(dice, ThreeOfAKind)
    if score != 22 { // 1+3+6+6+6
        t.Errorf("expected 22, got %d", score)
    }
}

func TestCalculateScore_ThreeOfAKind_NoMatch(t *testing.T) {
    dice := []int{1, 3, 5, 6, 6}
    score := calculateScore(dice, ThreeOfAKind)
    if score != 0 {
        t.Errorf("expected 0, got %d", score)
    }
}

func TestIsFullHouse(t *testing.T) {
    dice := []int{2, 2, 3, 3, 3}
    if !isFullHouse(dice) {
        t.Error("expected full house to be true")
    }
}
```

Run the tests:

```bash
go test ./...
```

Output:

```
ok      github.com/oglimmer/kniffel/model   0.003s
```

### Using testify for better assertions

While Go's built-in `testing` package works, the `testify` library provides more expressive assertions:

```bash
go get github.com/stretchr/testify
```

Now we can rewrite our tests more concisely:

```go
package model

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCalculateScore_ThreeOfAKind(t *testing.T) {
    dice := []int{1, 3, 6, 6, 6}
    score := calculateScore(dice, ThreeOfAKind)
    assert.Equal(t, 22, score)
}

func TestCalculateScore_ThreeOfAKind_NoMatch(t *testing.T) {
    dice := []int{1, 3, 5, 6, 6}
    score := calculateScore(dice, ThreeOfAKind)
    assert.Equal(t, 0, score)
}

func TestCalculateScore_Fives(t *testing.T) {
    dice := []int{1, 5, 5, 5, 2}
    score := calculateScore(dice, Fives)
    assert.Equal(t, 15, score)
}

func TestIsFullHouse_True(t *testing.T) {
    assert.True(t, isFullHouse([]int{2, 2, 3, 3, 3}))
}

func TestIsFullHouse_False(t *testing.T) {
    assert.False(t, isFullHouse([]int{2, 2, 3, 4, 3}))
}
```

### Table-driven tests

Go encourages "table-driven tests" - a pattern where you define test cases as data:

```go
func TestCalculateScore(t *testing.T) {
    tests := []struct {
        name     string
        dice     []int
        bt       BookingType
        expected int
    }{
        {"ones", []int{1, 1, 2, 3, 4}, Ones, 2},
        {"threes", []int{3, 3, 3, 1, 2}, Threes, 9},
        {"three of a kind", []int{1, 3, 6, 6, 6}, ThreeOfAKind, 22},
        {"no three of a kind", []int{1, 3, 5, 6, 6}, ThreeOfAKind, 0},
        {"full house", []int{2, 2, 3, 3, 3}, FullHouse, 25},
        {"kniffel", []int{5, 5, 5, 5, 5}, Kniffel, 50},
        {"chance", []int{1, 2, 3, 4, 5}, Chance, 15},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            score := calculateScore(tt.dice, tt.bt)
            assert.Equal(t, tt.expected, score)
        })
    }
}
```

This is a very common and idiomatic Go testing pattern. Each test case gets its own sub-test with a descriptive name.

# Step 5 - Integration tests

Integration tests test functionality beyond a single function. In our case we want to test the REST endpoints with a running server.

Go makes this easy with the `httptest` package from the standard library:

Create a file `handler/game_handler_test.go`:

```go
package handler

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/oglimmer/kniffel/model"
    "github.com/oglimmer/kniffel/service"
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

func TestCreateGame_Success(t *testing.T) {
    router := setupRouter()

    body := model.CreateGameRequest{PlayerNames: []string{"oli", "ilo"}}
    jsonBody, _ := json.Marshal(body)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/game/", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var response model.GameResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.GameID)
    assert.Equal(t, "oli", response.CurrentPlayerName)
    assert.Equal(t, 2, len(response.PlayerData))
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
```

Notice that Go's `httptest` package lets us test HTTP handlers **without starting a real server**. The `httptest.NewRecorder()` captures the response in memory. This is much faster than Spring's `@SpringBootTest` which starts the entire application context.

Run all tests:

```bash
go test ./...
```

To run with verbose output:

```bash
go test ./... -v
```

To run tests with coverage:

```bash
go test ./... -cover
```

# Step 6 - Containerization of our backend and frontend

## Writing a Dockerfile for Go

Go's single binary compilation makes Docker images incredibly small.

Inside the Go application directory, create a file `Dockerfile`:

```Dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /opt/build

# copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY . .

# run tests and build
RUN go test ./...
RUN CGO_ENABLED=0 go build -o kniffel .

# Runtime stage - using scratch or alpine for tiny images
FROM alpine:3.21

WORKDIR /opt/app

# copy the binary from the build stage
COPY --from=builder /opt/build/kniffel /opt/app/

# define how the container should run
CMD ["./kniffel"]
```

This is much simpler than the Java Dockerfile because:
* No JVM/JRE needed at runtime
* The final image only contains the binary and Alpine Linux (~10MB total vs ~200MB+ for Java)

Build and run:

```bash
docker build --tag kniffel-backend .
docker run --rm -p 8080:8080 kniffel-backend
```

## A Dockerfile for our Vue client

Same as the Java version:

```Dockerfile
FROM node:18 AS builder

COPY . /opt/frontend

RUN cd /opt/frontend && \
    npm ci && \
    npm run build

FROM nginx:stable-alpine

COPY --from=builder /opt/frontend/dist /usr/share/nginx/html

EXPOSE 80
```

Build and run:

```bash
docker build --tag kniffel-frontend .
docker run --rm -p 80:80 kniffel-frontend
```

## Putting it all together into a docker-compose.yml

Create a file `docker-compose.yml` in the parent directory:

```yml
services:
  backend:
    build: ./kniffel
    ports:
      - 8080:8080
  frontend:
    build: ./kniffel-client
    ports:
      - 80:80
```

Start both:

```bash
docker compose up --build
```

To run as a daemon:

```bash
docker compose up --build -d

# stop and remove containers with
docker compose down
```

# What we've learnt

* Go uses explicit error returns instead of exceptions - `if err != nil` is the core pattern
* Custom error types implement the `error` interface
* `errors.As` checks error types (like `instanceof` in Java)
* Go has built-in testing (`testing` package) - no JUnit needed
* `testify` provides more expressive assertions
* Table-driven tests are an idiomatic Go pattern
* `httptest` allows testing HTTP handlers without a running server
* Go Docker images are tiny because there's no runtime/VM
* Go concepts
    * `error` interface and custom error types
    * `errors.As` for type checking errors
    * `testing.T` for test functions
    * `httptest.NewRecorder` for HTTP testing
    * Table-driven test pattern

# Extras if you have time

* Read about all HTTP status codes https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
* Read more about Go error handling https://go.dev/blog/error-handling-and-go
* Read more about Go testing https://go.dev/doc/tutorial/add-a-test
* Read about testify https://github.com/stretchr/testify
