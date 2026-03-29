package internal

import (
	"github.com/armylong/armylong-go/internal/controllers"
	"github.com/armylong/go-library/service/command"
	"github.com/armylong/go-library/service/longgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func RegisterWeb(command command.BaseCommand) {
	command.AddCliCommand(&cobra.Command{
		Use:     "serve",
		Aliases: []string{"web"},
		Run: func(cmd *cobra.Command, args []string) {
			longgin.Start(func(engine *gin.Engine) {
				controllers.RegisterRouters(engine)
			})
		},
	})
}
