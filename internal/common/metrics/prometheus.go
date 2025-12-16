package metrics

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var dynamicCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dynamic_counter",
		Help: "Count custom keys",
	},
	[]string{"key"},
)

type PrometheusMetricsClientConfig struct {
	Host        string
	ServiceName string
}

type PrometheusMetricsClient struct {
	registry *prometheus.Registry
}

func NewPrometheusMetricsClient(config *PrometheusMetricsClientConfig) *PrometheusMetricsClient {
	client := &PrometheusMetricsClient{}
	client.initPrometheus(config)

	return client
}

func (c *PrometheusMetricsClient) Inc(key string, value int) {
	dynamicCounter.WithLabelValues(key).Add(float64(value))
}

func (c *PrometheusMetricsClient) initPrometheus(cfg *PrometheusMetricsClientConfig) {
	c.registry = prometheus.NewRegistry()
	c.registry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	// customer collectors:
	err := c.registry.Register(dynamicCounter)
	if err != nil {
		panic(fmt.Errorf("register dynamic_counter: %w", err))
	}

	// metadata
	prometheus.WrapRegistererWith(prometheus.Labels{
		"serviceName": cfg.ServiceName,
	}, c.registry)

	// export
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{}))
		if err := http.ListenAndServe(cfg.Host, metricsMux); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Str("addr", cfg.Host).Msg("prometheus metrics endpoint start failed")
		}
	}()
}
