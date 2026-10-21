package config

import (
	"math"
	"testing"
	"time"
)

func TestTelemetryRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*Telemetry)
	}{
		{"invalid URL", func(cfg *Telemetry) { cfg.Endpoint = ":invalid" }},
		{"credentials in URL", func(cfg *Telemetry) { cfg.Endpoint = "http://user:password@localhost:4318" }},
		{"invalid scheme", func(cfg *Telemetry) { cfg.Endpoint = "grpc://localhost:4317" }},
		{"query in URL", func(cfg *Telemetry) { cfg.Endpoint = "http://localhost:4318?key=secret" }},
		{"negative ratio", func(cfg *Telemetry) { cfg.SampleRatio = -1 }},
		{"excessive ratio", func(cfg *Telemetry) { cfg.SampleRatio = 2 }},
		{"NaN ratio", func(cfg *Telemetry) { cfg.SampleRatio = math.NaN() }},
		{"infinite ratio", func(cfg *Telemetry) { cfg.SampleRatio = math.Inf(1) }},
		{"invalid interval", func(cfg *Telemetry) { cfg.MetricInterval = 0 }},
		{"invalid level", func(cfg *Telemetry) { cfg.LogLevel = "verbose" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg := Telemetry{Enabled: true, Endpoint: "http://localhost:4318", SampleRatio: 1, MetricInterval: time.Second, LogLevel: "info"}
			test.change(&cfg)

			if cfg.Validate() == nil {
				t.Fatal("invalid telemetry configuration accepted")
			}
		})
	}
}
