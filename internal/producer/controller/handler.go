package controller

import (
	"encoding/json"
	"log"
	"os"

	"github.com/ImTheCurse/ConflowCI/internal/producer/provider/github"
	"github.com/gofiber/fiber/v2"
)

var logger = log.New(os.Stdout, "[Webhook Handler]: ", log.Llongfile|log.LstdFlags)

func HandleWebhook(ctx *fiber.Ctx) error {
	event := ctx.Get("X-GitHub-Event")
	body := ctx.Body()

	if event != "pull_request" {
		// dosen't mean much, we are sending back to the
		// github worker that sent us to the webhook
		logger.Printf("Invalid event type, expected pull_request, got: %v", event)
		return fiber.ErrBadRequest
	}
	var payload github.PullRequestPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		logger.Printf("Failed to unmarshal payload: %v", err)
		return fiber.ErrBadRequest
	}
	dir := "/tmp/conflowci"
	_, err := payload.ClonePullRequest("", dir)

	if err != nil {
		logger.Printf("Failed to clone repo: %v", err)
		return fiber.ErrBadRequest
	}

	return ctx.SendStatus(fiber.StatusOK)
}
