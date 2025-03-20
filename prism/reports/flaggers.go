package reports

import (
	"log/slog"
	"prism/prism/api"
	"prism/prism/openalex"
)

type WorkFlagger interface {
	Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string, authorName string) ([]api.Flag, error)

	Name() string

	DisableForUniversityReport() bool
}

type AuthorFlagger interface {
	Flag(logger *slog.Logger, authorName string) ([]api.Flag, error)

	Name() string
}
