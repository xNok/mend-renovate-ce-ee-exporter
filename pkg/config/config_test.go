package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNew simply double-check the default value are still what we think they are
func TestNew(t *testing.T) {
	c := NewValidConfig(t)
	assert.Equal(t, c, New())
}

// TestConfig_Validate simply check that the validation function is doing its job
func TestConfig_Validate(t *testing.T) {
	type fields struct {
		Log            Log
		OpenTelemetry  OpenTelemetry
		Server         Server
		Redis          Redis
		Scheduler      Scheduler
		Pull           Pull
		GarbageCollect GarbageCollect
	}
	tests := []struct {
		name    string
		gen     func(t *testing.T) Config
		wantErr bool
	}{
		{
			name: "OK - ValidConfig",
			gen:  NewValidConfig,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := tt.gen(t)
				if err := c.Validate(); (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func NewValidConfig(t *testing.T) Config {
	c := Config{}

	c.Log.Level = "info"
	c.Log.Format = "text"

	c.OpenTelemetry.GRPCEndpoint = ""

	c.Server.ListenAddress = ":8080"
	c.Server.Metrics.Enabled = true

	c.Pull.Metrics.OnInit = true
	c.Pull.Metrics.Scheduled = true
	c.Pull.Metrics.IntervalSeconds = 30

	c.GarbageCollect.Metrics.Scheduled = true
	c.GarbageCollect.Metrics.IntervalSeconds = 600

	return c
}
