package services

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

type AuthorReportUpdateNotifierData struct {
	EmailID string `json:"email_id"`
}

type AuthorReportUpdateNotifier struct {
	notifier *services.EmailMessenger
}

func NewAuthorReportUpdateNotifier(notifier *services.EmailMessenger) *AuthorReportUpdateNotifier {
	return &AuthorReportUpdateNotifier{
		notifier: notifier,
	}
}

func (h *AuthorReportUpdateNotifier) Validate(data []byte, interval int) error {
	if interval < 0 {
		return fmt.Errorf("interval must be greater than or equal to 0")
	}

	if interval > minUpdateInterval {
		return fmt.Errorf("interval must be less than or equal to %d seconds", minUpdateInterval)
	}

	var hookData AuthorReportUpdateNotifierData
	if err := json.Unmarshal(data, &hookData); err != nil {
		return fmt.Errorf("failed to unmarshal hook data: %w", err)
	}

	return nil
}

func (h *AuthorReportUpdateNotifier) Run(report api.Report, data []byte, lastRanAt time.Time) error {
	var hookData AuthorReportUpdateNotifierData
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

func (h *AuthorReportUpdateNotifier) Type() string {
	return "AuthorReportUpdateNotifier"
}

func (h *AuthorReportUpdateNotifier) CreateHookData(r *http.Request, payload []byte, interval int) (hookData []byte, err error) {
	email_id, err := auth.GetUserEmail(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get email id: %w", err)
	}
	obj := AuthorReportUpdateNotifierData{
		EmailID: email_id,
	}

	hookData, err = json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hook data: %w", err)
	}
	return hookData, nil
}

func (h *AuthorReportUpdateNotifier) notify(flags []api.Flag) error {
	if h.notifier == nil {
		return fmt.Errorf("notifier is not set")
	}
	// create the html content with the flags and send the email
	return nil
}
