package main

import (
	"flag"

	router "github.com/ImTheCurse/ConflowCI/internal/orchestrator/routes"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	grpcUtil "github.com/ImTheCurse/ConflowCI/pkg/grpc"
	"github.com/gofiber/fiber/v2"
)

func main() {
	grpcUtil.DefineFlags()
	configFilename := flag.String("config", "conflow-ci.yaml", "filename for config file.")
	flag.Parse()

	config.GetConfig(*configFilename)

	app := fiber.New()
	githubRouter := app.Group("/github")
	router.TaskRouter(githubRouter, *configFilename)

	app.Listen(":7777")

}
