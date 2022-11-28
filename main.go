package main

import (
	"os"

	"github.com/PlanckProject/go-commons/logger"
	"github.com/metamemelord/rev/commands"
)

func init() {
	logger.Configure(&logger.Config{Format: "text"}, nil)
	if os.Getenv("APP_MODE") == "release" {
		logger.SetLevel("info")
	}
}

func main() {
	commands.Run()
}
