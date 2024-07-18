package cacheable

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CacheRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: defaultMetricsPrefix,
		Name:      "cache_requests_total",
		Help:      "cache_requests_total",
	}, []string{"namespace"},
	)

	CacheHitTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: defaultMetricsPrefix,
		Name:      "cache_hit_total",
		Help:      "cache_hit_total",
	}, []string{"namespace"},
	)
)
