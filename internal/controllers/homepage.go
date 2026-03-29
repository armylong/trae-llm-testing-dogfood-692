package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func homepage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.gohtml", map[string]any{
		"user": nil,
	})
}
