package services

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"prism/prism/api"
	"prism/prism/reports"
	"prism/prism/schema"
	"prism/prism/services/auth"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Hook interface {
	Validate(data []byte, interval int) error

	Run(report api.Report, data []byte, lastRanAt time.Time) error

	CreateHookData(r *http.Request, payload []byte, interval int) (hookData []byte, err error)

	Type() string
}

type HookService struct {
	db    *gorm.DB
	hooks map[string]Hook

	stop chan struct{}
}

func NewHookService(db *gorm.DB, hooks map[string]Hook) HookService {
	return HookService{
		db:    db,
		hooks: hooks,
	}
}

func (s *HookService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/{report_id}", WrapRestHandler(s.CreateHook))

	return r
}

func (s *HookService) CreateHook(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		slog.Error("error getting user id", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reportId, err := URLParamUUID(r, "report_id")
	if err != nil {
		slog.Error("error parsing report id", "error", err)
		return nil, CodedError(err, http.StatusBadRequest)
	}

	params, err := ParseRequestBody[api.CreateHookRequest](r)
	if err != nil {
		slog.Error("error parsing request body", "error", err)
		return nil, CodedError(err, http.StatusBadRequest)
	}

	hook, ok := s.hooks[params.Action]
	fmt.Printf("params.Action: %v\n", params.Action)
	if !ok {
		slog.Error("invalid hook action", "action", params.Action)
		return nil, CodedError(errors.New("invalid hook action"), http.StatusUnprocessableEntity)
	}

	if err := hook.Validate(params.Data, params.Interval); err != nil {
		slog.Error("error validating hook data", "error", err)
		return nil, CodedError(err, http.StatusUnprocessableEntity)
	}

	if err := s.db.Transaction(func(txn *gorm.DB) error {
		var userReport schema.UserAuthorReport
		if err := txn.First(&userReport, "id = ?", reportId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return CodedError(reports.ErrReportNotFound, http.StatusNotFound)
			}
			slog.Error("error retrieving user author report", "error", err)
			return CodedError(reports.ErrReportAccessFailed, http.StatusInternalServerError)
		}

		if userReport.UserId != userId {
			return CodedError(reports.ErrUserCannotAccessReport, http.StatusForbidden)
		}

		hookData, err := hook.CreateHookData(r, params.Data, params.Interval)
		if err != nil {
			slog.Error("error creating hook data", "error", err)
			return CodedError(err, http.StatusInternalServerError)
		}

		hook := schema.AuthorReportHook{
			Id:           uuid.New(),
			UserReportId: reportId,
			Action:       params.Action,
			Data:         hookData,
			LastRanAt:    time.Now(),
			Interval:     params.Interval,
		}

		if err := txn.Create(&hook).Error; err != nil {
			slog.Error("error creating author report hook", "error", err)
			return CodedError(reports.ErrReportAccessFailed, http.StatusInternalServerError)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *HookService) RunNextHook() {
	err := s.db.Transaction(func(txn *gorm.DB) error {
		var userReports []schema.UserAuthorReport

		if err := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Limit(10).
			Preload("Report").
			Preload("Report.Flags").
			Preload("Hooks").
			Joins("JOIN author_reports ON author_reports.id = user_author_reports.report_id").
			Joins("JOIN author_report_hooks ON author_report_hooks.user_report_id = user_author_reports.id").
			Where(`NOW() > author_report_hooks.last_ran_at + (author_report_hooks.interval || ' seconds')::interval`).
			Find(&userReports).Error; err != nil {
			return fmt.Errorf("error retrieving reports with hooks to run: %w", err)
		}

		for _, report := range userReports {
			content, err := reports.ConvertReport(report)
			if err != nil {
				return fmt.Errorf("error converting report content: %w", err)
			}

			for _, hook := range report.Hooks {
				exec, ok := s.hooks[hook.Action]
				if !ok {
					return fmt.Errorf("invalid hook action: %s", hook.Action)
				}

				if err := exec.Run(content, hook.Data, hook.LastRanAt); err != nil {
					return fmt.Errorf("error running hook: %w", err)
				}

				if err := txn.Model(&hook).Update("last_ran_at", time.Now().UTC()).Error; err != nil {
					return fmt.Errorf("error updating hook last ran at: %w", err)
				}
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("error running hook", "error", err)
	}
}

func (s *HookService) RunHooks(checkInterval time.Duration) {
	s.stop = make(chan struct{})

	go func() {
		timer := time.Tick(checkInterval)

		for {
			select {
			case <-timer:
				s.RunNextHook()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *HookService) Stop() {
	if s.stop != nil {
		close(s.stop)
	}
}
