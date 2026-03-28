# Chapter 06

Database backend with Postgres

# Goal

We want to use a relational database - Postgres - as the storage backend.

# Context and Knowledge

* There are many different ways to integrate a database into a Go application. We will use **GORM** which is the most popular Go ORM (Object-Relational Mapper)
* GORM maps Go structs to database tables, similar to JPA/Hibernate in Java
* Unlike JPA which is a specification with multiple implementations (Hibernate, EclipseLink), GORM is a single library
* An alternative to GORM is `sqlx` which gives you more control with raw SQL, or even the standard library `database/sql`
* Logging is important for any server application. Go has `log/slog` in the standard library (since Go 1.21) for structured logging
* For testing with databases, we'll use `testcontainers-go` to run temporary Postgres containers

# Step 1 - Dependencies

Adding database support needs several steps. Let's do them one by one.

Install the needed dependencies:

```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

For testing with containers:

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

# Step 2 - Postgres Server

We need to start a Postgres server (or have access to one running on another server).

The easiest way is using Docker:

```bash
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres
```

For this tutorial you don't need to access the database server directly, but if you are interested you can use a client listed on https://wiki.postgresql.org/wiki/PostgreSQL_Clients. Download it and use the following parameters:

```
Host: localhost
Port: 5432 (standard)
User: postgres
Password: postgres
```

# Step 3 - Configuration

In Go, we typically use environment variables or a config struct for configuration. Create a file `config/config.go`:

```go
package config

import "os"

type Config struct {
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    ServerPort string
}

func Load() *Config {
    return &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", "postgres"),
        DBName:     getEnv("DB_NAME", "postgres"),
        ServerPort: getEnv("SERVER_PORT", "8080"),
    }
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
```

In Java/Spring you use `application.properties` with `${ENV_VAR:default}` syntax. In Go, we read environment variables directly. The pattern is the same: provide sensible defaults for local development but allow override via environment variables.

# Step 4 - GORM basics

GORM asks you to provide models (structs with special tags) that define the database schema.

## Defining models

Let's update our game models to work with GORM. Open `model/game.go` and modify the structs:

```go
import "gorm.io/gorm"

// KniffelGame is the main game entity
type KniffelGame struct {
    gorm.Model                          // adds ID, CreatedAt, UpdatedAt, DeletedAt
    GameID           string             `gorm:"uniqueIndex"`
    CurrentPlayerIdx int
    State            GameState
    RollRound        int
    Players          []KniffelPlayer    `gorm:"foreignKey:GameID;references:ID"`
    DiceRolls        []DiceRoll         `gorm:"foreignKey:GameID;references:ID"`
}

// KniffelPlayer is a player entity
type KniffelPlayer struct {
    gorm.Model
    GameID           uint               // foreign key to KniffelGame
    Name             string
    Score            int
    UsedBookingTypes []UsedBookingType  `gorm:"foreignKey:PlayerID;references:ID"`
}

// DiceRoll stores individual dice values (like @ElementCollection in JPA)
type DiceRoll struct {
    gorm.Model
    GameID uint
    Value  int
}

// UsedBookingType stores used booking types per player
type UsedBookingType struct {
    gorm.Model
    PlayerID    uint
    BookingType BookingType
}
```

Let's discuss the GORM specifics:

* `gorm.Model` - embeds ID (uint, primary key, auto-increment), CreatedAt, UpdatedAt, DeletedAt fields. This is similar to `@Id @GeneratedValue` in JPA.

* `gorm:"uniqueIndex"` - creates a unique index on this column, so we can efficiently look up games by their string ID.

* `gorm:"foreignKey:GameID;references:ID"` - defines the relationship. This is similar to `@OneToMany` in JPA.

* GORM automatically creates table names from struct names (snake_case): `KniffelGame` -> `kniffel_games`, `KniffelPlayer` -> `kniffel_players`.

### Comparison with JPA annotations

| JPA (Java) | GORM (Go) |
|------------|-----------|
| `@Entity` | struct with `gorm.Model` |
| `@Id @GeneratedValue` | `gorm.Model` (includes auto-increment ID) |
| `@OneToMany` | `gorm:"foreignKey:..."` |
| `@ManyToOne` | foreign key field (e.g., `GameID uint`) |
| `@Enumerated(EnumType.STRING)` | Store as string type |
| `@ElementCollection` | Separate struct + relationship |

## Database connection

Create a file `repository/database.go`:

```go
package repository

import (
    "fmt"
    "github.com/oglimmer/kniffel/config"
    "github.com/oglimmer/kniffel/model"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // auto-migrate creates/updates tables based on the struct definitions
    // similar to spring.jpa.hibernate.ddl-auto=update
    err = db.AutoMigrate(
        &model.KniffelGame{},
        &model.KniffelPlayer{},
        &model.DiceRoll{},
        &model.UsedBookingType{},
    )
    if err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    return db, nil
}
```

## Repository

Create a file `repository/game_repository.go`:

```go
package repository

import (
    "github.com/oglimmer/kniffel/model"
    "gorm.io/gorm"
)

type GameRepository struct {
    db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
    return &GameRepository{db: db}
}

func (r *GameRepository) Save(game *model.KniffelGame) error {
    return r.db.Save(game).Error
}

func (r *GameRepository) FindByGameID(gameID string) (*model.KniffelGame, error) {
    var game model.KniffelGame
    result := r.db.
        Preload("Players").
        Preload("Players.UsedBookingTypes").
        Preload("DiceRolls").
        Where("game_id = ?", gameID).
        First(&game)
    if result.Error != nil {
        return nil, result.Error
    }
    return &game, nil
}
```

### GORM repository methods vs JPA

| JPA Repository | GORM |
|---------------|------|
| `findAll()` | `db.Find(&games)` |
| `save(entity)` | `db.Save(&game)` |
| `findById(id)` | `db.First(&game, id)` |
| `deleteById(id)` | `db.Delete(&game, id)` |
| `findByGameId(gameId)` | `db.Where("game_id = ?", gameId).First(&game)` |

In JPA, you define an interface and Spring generates the implementation from method names. In GORM, you write the queries yourself using GORM's query builder. This is more explicit but gives you more control.

The `Preload` calls tell GORM to also load the related records (players, dice rolls) - similar to JPA's eager loading.

## Update GameService

Now update the `GameService` to use the repository instead of the in-memory map:

```go
type GameService struct {
    repo *repository.GameRepository
}

func NewGameService(repo *repository.GameRepository) *GameService {
    return &GameService{repo: repo}
}

func (s *GameService) CreateGame(playerNames []string) (*model.KniffelGame, error) {
    // validations...

    players := make([]model.KniffelPlayer, len(playerNames))
    for i, name := range playerNames {
        players[i] = model.KniffelPlayer{Name: name}
    }
    game := model.NewKniffelGame(players)

    if err := s.repo.Save(game); err != nil {
        return nil, fmt.Errorf("failed to save game: %w", err)
    }
    return game, nil
}

func (s *GameService) GetGameInfo(gameId string) (*model.KniffelGame, error) {
    game, err := s.repo.FindByGameID(gameId)
    if err != nil {
        return nil, &model.ErrNotFound{Message: gameId}
    }
    return game, nil
}
```

## Update main.go

Wire everything together:

```go
func main() {
    cfg := config.Load()

    db, err := repository.NewDatabase(cfg)
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }

    gameRepo := repository.NewGameRepository(db)
    gameService := service.NewGameService(gameRepo)
    gameHandler := handler.NewGameHandler(gameService)

    r := gin.Default()
    r.Use(cors.Default())

    gameGroup := r.Group("/api/v1/game")
    gameHandler.RegisterRoutes(gameGroup)

    r.Run(":" + cfg.ServerPort)
}
```

## Database transactions

What happens if saving the game fails after players are already created? We need database transactions.

In Go/GORM, transactions are explicit:

```go
func (s *GameService) CreateGame(playerNames []string) (*model.KniffelGame, error) {
    // ... validations ...

    err := s.repo.Transaction(func(tx *gorm.DB) error {
        // all operations inside this function are in one transaction
        // if any returns an error, everything is rolled back
        if err := tx.Create(&game).Error; err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return game, nil
}
```

In Spring, you use `@Transactional` on the class or method. In Go, you wrap the operations in a transaction callback. The effect is the same: either ALL database changes succeed, or NONE of them.

# Step 5 - Using a database in tests

We'll use `testcontainers-go` to run a temporary Postgres database during tests.

Create a file `repository/database_test.go`:

```go
package repository

import (
    "context"
    "fmt"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
    "github.com/oglimmer/kniffel/config"
)

func setupTestDB(t *testing.T) *config.Config {
    ctx := context.Background()

    pgContainer, err := postgres.Run(ctx,
        "postgres:latest",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForListeningPort("5432/tcp"),
        ),
    )
    if err != nil {
        t.Fatalf("failed to start postgres container: %v", err)
    }

    t.Cleanup(func() {
        pgContainer.Terminate(ctx)
    })

    host, _ := pgContainer.Host(ctx)
    port, _ := pgContainer.MappedPort(ctx, "5432")

    return &config.Config{
        DBHost:     host,
        DBPort:     port.Port(),
        DBUser:     "testuser",
        DBPassword: "testpass",
        DBName:     "testdb",
    }
}

func TestDatabaseConnection(t *testing.T) {
    cfg := setupTestDB(t)
    db, err := NewDatabase(cfg)
    assert.NoError(t, err)
    assert.NotNil(t, db)
}
```

This is similar to Spring's `@Testcontainers` with `@Container` and `@DynamicPropertySource`, but written explicitly. The `t.Cleanup` function ensures the container is stopped after the test - similar to JUnit's `@AfterAll`.

For integration tests that test the full HTTP flow with a database, you combine `httptest` with testcontainers:

```go
func TestCreateGame_WithDB(t *testing.T) {
    cfg := setupTestDB(t)
    db, _ := NewDatabase(cfg)

    gameRepo := NewGameRepository(db)
    gameService := service.NewGameService(gameRepo)
    gameHandler := handler.NewGameHandler(gameService)

    gin.SetMode(gin.TestMode)
    r := gin.Default()
    gameGroup := r.Group("/api/v1/game")
    gameHandler.RegisterRoutes(gameGroup)

    // ... use httptest as in chapter 05
}
```

# Step 6 - Logging

Go 1.21 introduced `log/slog` for structured logging in the standard library:

```go
import "log/slog"

func (g *KniffelGame) ReRollDice(diceToKeep []int) {
    slog.Debug("rolling dice",
        "gameId", g.GameID,
        "currentDice", g.DiceRolls,
        "toKeep", diceToKeep,
    )
    // ... rest of the method
}
```

The `slog` package provides structured logging with key-value pairs:

* `slog.Debug()` - debug level
* `slog.Info()` - info level
* `slog.Warn()` - warning level
* `slog.Error()` - error level

### Configuring log levels

Set the log level at startup:

```go
func main() {
    // set log level based on environment
    logLevel := slog.LevelInfo
    if os.Getenv("LOG_LEVEL") == "debug" {
        logLevel = slog.LevelDebug
    }

    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: logLevel,
    }))
    slog.SetDefault(logger)

    // ... rest of main
}
```

### GORM SQL logging

To see SQL queries GORM executes:

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // shows all SQL
})
```

# Step 7 - Health check endpoint

In Spring you added `spring-boot-starter-actuator` for a health endpoint. In Go, we write it ourselves (it's just a few lines):

```go
r.GET("/health", func(c *gin.Context) {
    // check database connection
    sqlDB, err := db.DB()
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
        return
    }
    if err := sqlDB.Ping(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "UP"})
})
```

# Step 8 - Extending Docker Compose for database backend

```yml
version: '3'

services:
  db:
    image: postgres
    environment:
      POSTGRES_PASSWORD: postgres
      POSTGRES_USER: postgres
      POSTGRES_DB: postgres
    volumes:
      - ./data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready --user postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 15s
  backend:
    build: ./kniffel
    ports:
      - 8080:8080
    environment:
      DB_HOST: db
    depends_on:
      db:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 5s
  frontend:
    build: ./kniffel-client
    ports:
      - 80:80
    depends_on:
      backend:
        condition: service_healthy
```

Note: We use `wget` instead of `curl` in the health check because our Alpine-based Go image doesn't include curl by default. Alternatively, you could add curl to the Dockerfile or use a Go-based health check binary.

The startup time for the Go backend is typically under 1 second (vs several seconds for Spring Boot), so the `start_period` can be much shorter.

# Step 9 - Deploying (the Vue UI) to production

Same as the Java version - we need to make the REST API URL configurable.

Update `vite.config.ts`:

```ts
export default defineConfig({
  define: {
    __API_URL__: JSON.stringify(process.env.API_URL ?? "http://localhost:8080")
  },
  // ... rest
})
```

Update `env.d.ts`:

```ts
/// <reference types="vite/client" />
declare const __API_URL__: string;
```

Update `App.vue`:

```ts
const client = createClient<paths>({ baseUrl: `${__API_URL__}` });
```

Update the frontend Dockerfile:

```Dockerfile
FROM node:18 AS builder

ARG API_URL

COPY . /opt/frontend

RUN cd /opt/frontend && \
    npm ci && \
    npm run build

FROM nginx:stable-alpine

COPY --from=builder /opt/frontend/dist /usr/share/nginx/html

EXPOSE 80
```

Build with a custom API URL:

```bash
docker build --tag kniffel-frontend --build-arg API_URL=https://api-kniffel.oglimmer.com .
```

# What we've learnt

* What is an ORM and how to use GORM in Go
* How to add Postgres to a Go application
* GORM model tags (`gorm.Model`, `gorm:"foreignKey:..."`, etc.)
* How to write repositories with GORM's query builder
* Database transactions in Go (explicit callback vs Spring's `@Transactional`)
* Structured logging with `log/slog`
* Writing a health check endpoint
* Using testcontainers-go for database tests
* Docker Compose with health check dependencies
* Build time variables in Vue
* GORM concepts
    * `gorm.Model` - base model with ID, timestamps
    * `db.AutoMigrate` - auto-create/update tables
    * `db.Create`, `db.Save`, `db.First`, `db.Find` - CRUD operations
    * `db.Where`, `db.Preload` - queries and eager loading
    * `db.Transaction` - explicit transactions

# Extras if you have time

* Read the GORM documentation https://gorm.io/docs/
* Read about `log/slog` https://pkg.go.dev/log/slog
* Read about testcontainers-go https://golang.testcontainers.org/
* Explore `sqlx` as a lighter alternative to GORM https://github.com/jmoiron/sqlx
