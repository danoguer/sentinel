package metrics

import (
	"net/http"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	LinesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sentinel_lines_processed_total",
		Help: "The total number of shell log lines processed by Sentinel",
	})

	SecretsRedacted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sentinel_secrets_redacted_total",
		Help: "The total number of sensitive secrets scrubbed by the regex engine",
	})

	SanitizationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "sentinel_sanitization_duration_seconds",
		Help:    "The duration of the regex sanitization process in seconds",
		Buckets: prometheus.DefBuckets, // Buckets nativos de Prometheus (de microsegundos a segundos)
	})
)

func StartServer(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("🛡️ Sentinel Metrics server listening on %s/metrics", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("❌ Metrics server failed: %v", err)
		}
	}()
}
