package main

import (
	router "github.com/ImTheCurse/ConflowCI/internal/producer/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()
	githubRouter := app.Group("/github")
	router.TaskRouter(githubRouter)

	app.Listen(":7777")

}
