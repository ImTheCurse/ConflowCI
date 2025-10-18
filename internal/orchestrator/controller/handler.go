package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ImTheCurse/ConflowCI/internal/provider/github"
	"github.com/ImTheCurse/ConflowCI/internal/sync"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/gofiber/fiber/v2"
)

var logger = log.New(os.Stdout, "[Webhook Handler]: ", log.Lshortfile|log.LstdFlags)

// TODO: check for private repo and token.
func HandleWebhook(ctx *fiber.Ctx, cfg config.ValidatedConfig) error {
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

	// "pull/6/head:pr-6"
	refSpec := fmt.Sprintf("pull/%d/head:pr-%d", payload.Number, payload.PullRequest.ID)
	wb := sync.NewWorkerBuilder(cfg, "origin", payload.PullRequest.OriginBranch.Ref, refSpec)
	outputs := wb.BuildAllEndpoints()

	logger.Printf("Outputs: %v", outputs)

	return ctx.SendStatus(fiber.StatusOK)
}
