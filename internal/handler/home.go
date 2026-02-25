package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Version":     0.0001,
		"title":       "Upatu",
		"description": "Rotational Group Savings & Messaging Platform",
	})
}
