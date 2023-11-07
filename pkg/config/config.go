package config

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var validate *validator.Validate

type Config struct {
	// Log configuration for the exporter
	Log Log `yaml:"log"`

	// OpenTelemetry configuration
	OpenTelemetry OpenTelemetry `yaml:"opentelemetry"`

	// Server related configuration
	Server Server `yaml:"server"`

	// Redis related configuration
	Redis Redis `yaml:"redis"`

	// Scheduler related configuration
	Scheduler Scheduler `yaml:"scheduler"`

	// Pull configuration
	Pull Pull `yaml:"pull"`

	// GarbageCollect configuration
	GarbageCollect GarbageCollect `yaml:"garbage_collect"`
}

// Validate will throw an error if the Config parameters are whether incomplete or incorrect.
func (c Config) Validate() error {
	if validate == nil {
		validate = validator.New()
	}

	return validate.Struct(c)
}

// Server ..
type Server struct {
	// Enable profiling pages
	EnablePprof bool `default:"false" yaml:"enable_pprof"`

	// [address:port] to make the process listen upon
	ListenAddress string `default:":8080" yaml:"listen_address"`

	Metrics ServerMetrics `yaml:"metrics"`
	Webhook ServerWebhook `yaml:"webhook"`
}

// ServerMetrics ..
type ServerMetrics struct {
	// Enable /metrics endpoint
	Enabled bool `default:"true" yaml:"enabled"`

	// Enable OpenMetrics content encoding in prometheus HTTP handler
	EnableOpenmetricsEncoding bool `default:"false" yaml:"enable_openmetrics_encoding"`
}

// ServerWebhook ..
type ServerWebhook struct {
	// Enable /webhook endpoint to support webhook requests
	Enabled bool `default:"false" yaml:"enabled"`

	// Secret token to authenticate legitimate webhook requests coming from the GitLab server
	SecretToken string `validate:"required_if=Enabled true" yaml:"secret_token"`
}

// Log holds runtime logging configuration.
type Log struct {
	// Log level
	Level string `default:"info" validate:"required,oneof=trace debug info warning error fatal panic"`

	// Log format
	Format string `default:"text" validate:"oneof=text json"`
}

// OpenTelemetry related configuration.
type OpenTelemetry struct {
	// gRPC endpoint of the opentelemetry collector
	GRPCEndpoint   string `yaml:"grpc_endpoint"`
	ServiceNameKey string `yaml:"service_name_key"`
}

// Redis ..
type Redis struct {
	// URL used to connect onto the redis endpoint
	// format: redis[s]://[:password@]host[:port][/db-number][?option=value])
	URL string `yaml:"url"`
}

// Scheduler ..
type Scheduler struct {
	// BufferSize for the task/job queue
	MaximumJobsQueueSize int `yaml:"maximum_jobs_queue_size"`
}

// SchedulerConfig ..
type SchedulerConfig struct {
	OnInit          bool
	Scheduled       bool
	IntervalSeconds int
}

// Log returns some logging fields to showcase the configuration to the enduser.
func (sc SchedulerConfig) Log() log.Fields {
	onInit, scheduled := "no", "no"
	if sc.OnInit {
		onInit = "yes"
	}

	if sc.Scheduled {
		scheduled = fmt.Sprintf("every %vs", sc.IntervalSeconds)
	}

	return log.Fields{
		"on-init":   onInit,
		"scheduled": scheduled,
	}
}

// Pull ..
type Pull struct {
	// Metrics configuration
	Metrics struct {
		OnInit          bool `default:"true" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"30" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"metrics"`
}

// GarbageCollect ..
type GarbageCollect struct {
	// Metrics configuration
	Metrics struct {
		OnInit          bool `default:"false" yaml:"on_init"`
		Scheduled       bool `default:"true" yaml:"scheduled"`
		IntervalSeconds int  `default:"600" validate:"gte=1" yaml:"interval_seconds"`
	} `yaml:"metrics"`
}

// New returns a new config with the default parameters.
func New() (c Config) {
	defaults.MustSet(&c)

	return
}

// ToYAML ..
func (c Config) ToYAML() string {
	c.Server.Webhook.SecretToken = "*******"

	b, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}

	return string(b)
}
