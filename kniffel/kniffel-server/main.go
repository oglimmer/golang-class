package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oglimmer/kniffel/config"
	"github.com/oglimmer/kniffel/handler"
	"github.com/oglimmer/kniffel/repository"
	"github.com/oglimmer/kniffel/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/oglimmer/kniffel/docs"
)

// @title Kniffel Game API
// @version 0.0.1
// @description Kniffel as a service - KaaS
// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0

// @externalDocs.description Kniffel Regeln
// @externalDocs.url https://www.schmidtspiele.de/files/Produkte/4/49030%20-%20Kniffel/49203_49030_Kniffel_DE.pdf
func main() {
	cfg := config.Load()

	// Configure structured logging
	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Wire dependencies - try database, fall back to in-memory
	var gameService *service.GameService
	var gameRepo *repository.GameRepository

	db, err := repository.NewDatabase(cfg)
	if err != nil {
		slog.Warn("failed to connect to database, using in-memory storage", "error", err)
		gameService = service.NewGameService()
	} else {
		slog.Info("connected to database", "host", cfg.DBHost, "port", cfg.DBPort)
		gameRepo = repository.NewGameRepository(db)
		gameService = service.NewGameServiceWithRepo(gameRepo)
	}

	gameHandler := handler.NewGameHandler(gameService)
	serverHandler := handler.NewServerHandler()

	r := gin.Default()
	r.Use(cors.Default())

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		if gameRepo != nil {
			sqlDB, err := gameRepo.DB().DB()
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
				return
			}
			if err := sqlDB.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Register routes
	serverGroup := r.Group("/api/v1/server")
	serverHandler.RegisterRoutes(serverGroup)

	gameGroup := r.Group("/api/v1/game")
	gameHandler.RegisterRoutes(gameGroup)

	slog.Info("starting server", "port", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
