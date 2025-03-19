package monitoring

import "github.com/prometheus/client_golang/prometheus"

var (
	OpenalexCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openalex_calls",
		Help: "Total calls made to OpenAlex",
	}, []string{"status"})

	SerpapiCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "serpai_calls",
		Help: "Total calls made to serpapi",
	}, []string{"status"})
)
