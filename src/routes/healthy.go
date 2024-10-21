package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Show the Health status of server.
// @Description get the Health status of server.
// @Tags Health Status
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func HealthCheck(g *gin.Context) {
	g.JSON(http.StatusOK, gin.H{"msg": "server is healthy"})
}

// func HttpErrorHandler(err error, c echo.Context) {
// 	fmt.Println(err)
// }
