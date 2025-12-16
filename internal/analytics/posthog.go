package analytics

import (
	"os"
	"runtime"

	"github.com/posthog/posthog-go"
)

type Client struct {
	ph          posthog.Client
	distinctID  string
	userMachine string
}

func NewFromEnv() *Client {
	apiKey := os.Getenv("POSTHOG_API_KEY")
	if apiKey == "" {
		return nil
	}

	endpoint := os.Getenv("POSTHOG_ENDPOINT")
	if endpoint == "" {
		// Default to US ingest. Override via POSTHOG_ENDPOINT for EU/self-hosted.
		endpoint = "https://us.i.posthog.com"
	}

	ph, err := posthog.NewWithConfig(apiKey, posthog.Config{Endpoint: endpoint})
	if err != nil {
		return nil
	}

	host, _ := os.Hostname()
	if host == "" {
		host = "unknown"
	}

	userMachine := "architecture_" + runtime.GOARCH

	return &Client{
		ph:          ph,
		distinctID:  host,
		userMachine: userMachine,
	}
}

func (c *Client) Close() {
	if c == nil || c.ph == nil {
		return
	}
	_ = c.ph.Close()
}

func (c *Client) StepCompleted(workflowName, stepID, stepType string, durationMs int64) {
	if c == nil || c.ph == nil {
		return
	}

	props := posthog.NewProperties().
		Set("workflow_name", workflowName).
		Set("step_id", stepID).
		Set("step_type", stepType).
		Set("duration_ms", durationMs).
		Set("user_machine", c.userMachine)

	c.ph.Enqueue(posthog.Capture{
		DistinctId: c.distinctID,
		Event:      "step_completed",
		Properties: props,
	})
}
