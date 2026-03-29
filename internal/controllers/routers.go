package controllers

import (
	"errors"
	"net/http"

	"github.com/armylong/armylong-go/internal/controllers/feishu"
	"github.com/armylong/armylong-go/internal/controllers/user"
	"github.com/armylong/armylong-go/internal/controllers/yangfen"

	"github.com/armylong/go-library/service/longgin"
	"github.com/gin-gonic/gin"
)

func RegisterRouters(engine *gin.Engine) {

	engine.LoadHTMLGlob(`./templates/*.gohtml`)

	engine.Static("/static", "./static")

	engine.Any(`/`, homepage)

	engine.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, longgin.ErrorWithContext(ctx, errors.New("not found"), longgin.ErrorNotFound))
	})

	userRoot := engine.Group("/user")
	longgin.RegisterJsonController(userRoot.Group("/demo"), &user.DemoController{})

	yangfenRoot := engine.Group("/yangfen")
	longgin.RegisterJsonController(yangfenRoot, &yangfen.YangfenController{})

	feishuRoot := engine.Group("/feishu")
	longgin.RegisterJsonController(feishuRoot, &feishu.FeishuCloudDocController{})

}
