package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFileInvalidPath(t *testing.T) {
	cfg, err := ParseFile("/path_do_not_exist.yml")
	assert.Error(t, err)
	assert.Equal(t, Config{}, cfg)
}

func TestParseInvalidYaml(t *testing.T) {
	cfg, err := Parse(FormatYAML, []byte("invalid_yaml"))
	assert.Error(t, err)
	assert.Equal(t, Config{}, cfg)
}

func TestParseValidYaml(t *testing.T) {
	yamlConfig, err := os.ReadFile(filepath.Clean("./testdata/ValidConfig.yaml"))
	assert.NoError(t, err)

	cfg, err := Parse(FormatYAML, []byte(yamlConfig))
	assert.NoError(t, err)

	// Create a config from the default values
	xcfg := New()
	xcfg.Log.Level = "trace"
	xcfg.Log.Format = "json"

	xcfg.OpenTelemetry.GRPCEndpoint = "otlp-collector:4317"
	xcfg.OpenTelemetry.ServiceNameKey = "mend-renovate-ce-ee-exporter"

	xcfg.Server.EnablePprof = true
	xcfg.Server.ListenAddress = ":1025"
	xcfg.Server.Metrics.Enabled = false
	xcfg.Server.Metrics.EnableOpenmetricsEncoding = false
	xcfg.Server.Webhook.Enabled = true
	xcfg.Server.Webhook.SecretToken = "secret"

	xcfg.Redis.URL = "redis://popopo:1337"

	xcfg.Pull.Metrics.OnInit = false
	xcfg.Pull.Metrics.Scheduled = false
	xcfg.Pull.Metrics.IntervalSeconds = 4

	xcfg.GarbageCollect.Metrics.OnInit = true
	xcfg.GarbageCollect.Metrics.Scheduled = false
	xcfg.GarbageCollect.Metrics.IntervalSeconds = 4

	// Test variable assignments
	assert.Equal(t, xcfg, cfg)
}
