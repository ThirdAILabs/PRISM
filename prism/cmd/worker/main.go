package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"prism/cmd"
	"prism/reports"
	"prism/reports/flaggers"
	"time"
)

type Config struct {
	PostgresUri string `yaml:"postgres_uri"`
	Logfile     string
	NdbLicense  string `yaml:"ndb_license"`
}

func (c *Config) logfile() string {
	if c.Logfile == "" {
		return "prism_backend.log"
	}
	return c.Logfile
}

func main() {
	var config Config
	cmd.LoadConfig(&config)

	logFile, err := os.OpenFile(config.logfile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	cmd.InitLogging(logFile)

	opts := flaggers.ReportProcessorOptions{}

	processor, err := flaggers.NewReportProcessor(opts)
	if err != nil {
		log.Fatalf("error creating work processor: %w", err)
	}

	db := cmd.InitDb(config.PostgresUri)

	reportManager := reports.NewManager(db)

	for {
		time.Sleep(10 * time.Second)

		nextReport, err := reportManager.GetNextReport()
		if err != nil {
			slog.Error("error checking for next report", "error", err)
			continue
		}
		if nextReport == nil {
			continue
		}

		flags, err := processor.ProcessReport(*nextReport)
		if err != nil {
			slog.Error("error processing report: %w")

			if err := reportManager.UpdateReport(nextReport.Id, "failed", []byte(err.Error())); err != nil {
				slog.Error("error updating report status to failed", "error", err)
			}
			continue
		}

		content, err := reports.FormatReport(flags)
		if err != nil {
			slog.Error("error formatting report: %w", err)

			if err := reportManager.UpdateReport(nextReport.Id, "failed", []byte(err.Error())); err != nil {
				slog.Error("error updating report status to failed", "error", err)
			}
			continue
		}

		contentBytes, err := json.Marshal(content)
		if err != nil {
			slog.Error("error serializing report: %w", err)

			if err := reportManager.UpdateReport(nextReport.Id, "failed", []byte(err.Error())); err != nil {
				slog.Error("error updating report status to failed", "error", err)
			}
			continue
		}

		if err := reportManager.UpdateReport(nextReport.Id, "complete", contentBytes); err != nil {
			slog.Error("error updating report status to complete", "error", err)
		}
	}
}
