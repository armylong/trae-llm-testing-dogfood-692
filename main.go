/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/armylong/armylong-go/internal"
	"github.com/armylong/go-library/service/command"
)

func main() {
	command.Go(func(command command.BaseCommand) {
		internal.RegisterCmd(command)
		internal.RegisterWeb(command)
	})
}
