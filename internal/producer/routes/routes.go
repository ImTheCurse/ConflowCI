package routes

import (
	"github.com/ImTheCurse/ConflowCI/internal/producer/controller"
	"github.com/gofiber/fiber/v2"
)

func TaskRouter(router fiber.Router) {
	router.Post("webhook", controller.HandleWebhook)
}
