package observability

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIdentityEnvBeatsApsorc(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".apsorc"), []byte(`{"serviceName":"from-file","serviceId":"1","workspaceId":"9"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("APP_NAME", "")
	t.Setenv("APSO_SERVICE_ID", "")
	t.Setenv("SERVICE_ID", "")
	t.Setenv("APSO_WORKSPACE_ID", "")
	if got := ServiceName(); got != "from-file" {
		t.Fatalf("service name = %q", got)
	}
	got := map[string]string{}
	for _, kv := range identityAttributes() {
		got[string(kv.Key)] = kv.Value.AsString()
	}
	if got["apso.service.id"] != "1" || got["apso.workspace.id"] != "9" {
		t.Fatalf("from file: %v", got)
	}
	t.Setenv("APSO_SERVICE_ID", "42")
	got = map[string]string{}
	for _, kv := range identityAttributes() {
		got[string(kv.Key)] = kv.Value.AsString()
	}
	if got["apso.service.id"] != "42" {
		t.Fatalf("env should win: %v", got)
	}
}

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
