package routes

import (
	"log"
	"os"

	"github.com/ImTheCurse/ConflowCI/internal/orchestrator/controller"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/gofiber/fiber/v2"
)

var logger = log.New(os.Stdout, "[Routes]: ", log.Lshortfile|log.LstdFlags)

func TaskRouter(router fiber.Router, filename string) {
	router.Post("webhook", func(c *fiber.Ctx) error {
		logger.Println("Getting config...")
		cfg, err := config.GetConfig(filename)
		if err != nil {
			logger.Printf("Error getting config, got: %v", err)
			return err
		}
		return controller.HandleWebhook(c, *cfg)
	})
}
