package monitoring

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "http_requests",
		Help: "Total number of HTTP requests",
	}, []string{"method", "route", "status"})
)

func HandlerMetrics(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		// This is so that GET /api/v1/reports/abc is formatted as GET /api/v1/reports/{report_id}
		rctx := chi.RouteContext(r.Context())
		routePattern := strings.Replace(strings.Join(rctx.RoutePatterns, ""), "/*/", "/", -1)

		requestSummary.WithLabelValues(r.Method, routePattern, strconv.Itoa(ww.Status())).Observe(float64(time.Since(start).Milliseconds()))
	}
	return http.HandlerFunc(fn)
}

var (
	AuthorReportsCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "author_reports_created",
		Help: "Total number of author reports created",
	}, []string{"organization"})

	AuthorReportsFoundInCache = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "author_reports_found_in_cache",
		Help: "Total number of author reports found in cache",
	}, []string{"organization"})

	UniAuthorReportsCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uni_author_reports_created",
		Help: "Total number of university author reports created",
	}, []string{"organization"})

	UniAuthorReportsFoundInCache = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uni_author_reports_found_in_cache",
		Help: "Total number of university author reports found in cache",
	}, []string{"organization"})

	UniReportsCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uni_reports_created",
		Help: "Total number of university reports created",
	}, []string{"organization"})

	UniReportsFoundInCache = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uni_reports_found_in_cache",
		Help: "Total number of university reports found in cache",
	}, []string{"organization"})
)

func ExposeBackendMetrics(port int) {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		requestSummary,
		AuthorReportsCreated,
		AuthorReportsFoundInCache,
		UniAuthorReportsCreated,
		UniAuthorReportsFoundInCache,
		UniReportsCreated,
		UniReportsFoundInCache,
	)

	slog.Info("exposing backend metrics", "port", port)

	go func() {
		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
			log.Fatalf("error starting metrics server: %v", err)
		}
	}()
}
