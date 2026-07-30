package observability

import "testing"

func TestEnabled(t *testing.T) {
	t.Setenv("OTEL_ENABLED", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("APSO_TELEMETRY", "")
	if Enabled() {
		t.Fatal("must be off by default")
	}
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4318")
	if !Enabled() {
		t.Fatal("endpoint should enable")
	}
	t.Setenv("APSO_TELEMETRY", "off")
	if Enabled() {
		t.Fatal("APSO_TELEMETRY=off must win")
	}
}
