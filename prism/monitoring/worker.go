package monitoring

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ReportsProcessed = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "reports_processed",
		Help: "Total reports processed",
	})

	TotalFlags = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "total_flags",
		Help: "Total flags generated",
	}, []string{"flag_type"})

	FlaggerErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "flagger_errors",
		Help: "Total flagger errors",
	}, []string{"flagger"})

	ReportUpdateErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "report_update_errors",
		Help: "Total report update errors",
	})

	GrobidCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grobid_calls",
		Help: "Total calls made to grobid",
	}, []string{"status"})

	PdfCacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pdf_cache_hits",
		Help: "Total number of pdf cache hits",
	})

	PdfCacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pdf_cache_hits",
		Help: "Total number of pdf cache misses",
	})

	PdfCacheErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pdf_cache_errors",
		Help: "Total number of pdf cache errors",
	})

	PdfCacheUploadErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pdf_cache_upload_errors",
		Help: "Total number of pdf cache upload errors",
	})

	HttpDownloadSuccesses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_download_successes",
		Help: "Total number of successful http downloads",
	})

	HttpDownloadErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_download_errors",
		Help: "Total number of failed http downloads",
	})

	PlaywrightDownloadSuccesses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "playwright_download_successes",
		Help: "Total number of successful playwright downloads",
	})

	PlaywrightDownloadErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "playwright_download_errors",
		Help: "Total number of failed playwright downloads",
	})
)

func WorkRegistry() *prometheus.Registry {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		GrobidCalls,
	)

	return registry
}

func ExposeWorkerMetrics(port int) {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		ReportsProcessed,
		TotalFlags,
		FlaggerErrors,
		ReportUpdateErrors,
		GrobidCalls,
		PdfCacheHits,
		PdfCacheMisses,
		PdfCacheErrors,
		PdfCacheUploadErrors,
		HttpDownloadSuccesses,
		HttpDownloadErrors,
		PlaywrightDownloadSuccesses,
		PlaywrightDownloadErrors,
	)

	slog.Info("exposing worker metrics", "port", port)

	go func() {
		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
			log.Fatalf("error starting metrics server: %v", err)
		}
	}()
}
