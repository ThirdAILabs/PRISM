package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prism/prism/api"
	"prism/prism/services"
	"prism/prism/services/auth"
	"time"
)

var minUpdateInterval = 7 * 24 * 60 * 60 // 7 days in seconds

type FlagTrackerData struct {
	EmailID string `json:"email_id"`
}

type FlagTracker struct {
	notifier *services.EmailNotifier
}

func NewFlagTrackerHook(notifier *services.EmailNotifier) *FlagTracker {
	return &FlagTracker{
		notifier: notifier,
	}
}

func (h *FlagTracker) Validate(data []byte, interval int) error {
	if interval < 0 {
		return fmt.Errorf("interval must be greater than or equal to 0")
	}

	if interval > minUpdateInterval {
		return fmt.Errorf("interval must be less than or equal to %d seconds", minUpdateInterval)
	}

	var hookData FlagTrackerData
	if err := json.Unmarshal(data, &hookData); err != nil {
		return fmt.Errorf("failed to unmarshal hook data: %w", err)
	}

	return nil
}

func (h *FlagTracker) Run(report api.Report, data []byte, lastRanAt time.Time) error {
	var hookData FlagTrackerData
	if err := json.Unmarshal(data, &hookData); err != nil {
		return fmt.Errorf("failed to unmarshal hook data: %w", err)
	}

	updatedFlags := make([]api.Flag, 0)
	for _, flags := range report.Content {
		for _, flag := range flags {
			if flag.GetLastUpdatedAt().After(lastRanAt) {
				updatedFlags = append(updatedFlags, flag)
			}
		}
	}

	return h.notify(updatedFlags)
}

func (h *FlagTracker) Type() string {
	return "FlagTracker"
}

func (h *FlagTracker) CreateHookData(r *http.Request, params api.CreateHookRequest) (hookData []byte, err error) {
	email_id, err := auth.GetEmailId(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get email id: %w", err)
	}
	obj := FlagTrackerData{
		EmailID: email_id,
	}

	hookData, err = json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hook data: %w", err)
	}
	return hookData, nil
}

func (h *FlagTracker) notify(flags []api.Flag) error {
	// create the html content with the flags and send the email
	return nil
}
