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

	refSpec := fmt.Sprintf("pull/%d/head:pr-%d", payload.Number, payload.PullRequest.ID)
	wb := sync.NewWorkerBuilder(cfg, "origin", payload.PullRequest.OriginBranch.Ref, refSpec)
	outputs := wb.BuildAllEndpoints()

	for _, job := range cfg.Pipeline.Tasks {
		logger.Printf("Running task: %s", job.Name)
		te, err := sync.NewTaskExecutor(cfg, job, wb.Name)
		if err != nil {
			logger.Printf("Failed to create task executor for task: %s", job.Name)
			continue
		}
		err = te.RunTaskOnAllMachines()
		if err != nil {
			logger.Printf("Failed to run task: %s", job.Name)
		}
		logger.Printf("%s runner output: %v", job.Name, te.Outputs)
	}
	errs := wb.RemoveAllRepositoryWorkspaces()

	logger.Printf("RemoveAllRepositoryWorkspaces errors: %v", errs)
	logger.Printf("Build Outputs: %v", outputs)

	return ctx.SendStatus(fiber.StatusOK)
}
