package handler

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// ServerStatus is the response for the server status endpoint
type ServerStatus struct {
	ServerTime    int64  `json:"serverTime"`
	ServerVersion string `json:"serverVersion"`
	ServerName    string `json:"serverName"`
	ServerStatus  string `json:"serverStatus"`
}

// ServerHandler handles server-related HTTP requests
type ServerHandler struct{}

// NewServerHandler creates a new ServerHandler
func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}

// RegisterRoutes registers all server routes
func (h *ServerHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/goroutine-dump", h.GoroutineDump)
	r.GET("/status", h.Status)
}

// GoroutineDump dumps all goroutine stack traces
// @Summary Dump goroutine stack traces
// @Description Prints all goroutine stack traces to stdout
// @Tags server
// @Success 200
// @Router /api/v1/server/goroutine-dump [post]
func (h *ServerHandler) GoroutineDump(c *gin.Context) {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	fmt.Println(string(buf[:n]))
	c.Status(http.StatusOK)
}

// Status returns the server status
// @Summary Get server status
// @Description Returns server time, version and status
// @Tags server
// @Produce json
// @Success 200 {object} ServerStatus
// @Router /api/v1/server/status [get]
func (h *ServerHandler) Status(c *gin.Context) {
	status := ServerStatus{
		ServerTime:    time.Now().UnixMilli(),
		ServerVersion: "1.0.0",
		ServerName:    "Kniffel Server",
		ServerStatus:  "OK",
	}
	c.JSON(http.StatusOK, status)
}
