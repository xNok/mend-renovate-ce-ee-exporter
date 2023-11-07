package cmd

import (
	stdlibLog "log"
	"os"
	"time"

	"github.com/go-logr/stdr"
	"github.com/mvisonneau/go-helpers/logger"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"github.com/urfave/cli/v2"
	"github.com/vmihailenco/taskq/v4"

	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/config"
)

var start time.Time

func configure(ctx *cli.Context) (cfg config.Config, err error) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	assertStringVariableDefined(ctx, "config")

	cfg, err = config.ParseFile(ctx.String("config"))
	if err != nil {
		return
	}

	configCliOverrides(ctx, &cfg)

	if err = cfg.Validate(); err != nil {
		return
	}

	// Configure logger
	if err = logger.Configure(
		logger.Config{
			Level:  cfg.Log.Level,
			Format: cfg.Log.Format,
		},
	); err != nil {
		return
	}

	log.AddHook(
		otellogrus.NewHook(
			otellogrus.WithLevels(
				log.PanicLevel,
				log.FatalLevel,
				log.ErrorLevel,
				log.WarnLevel,
			),
		),
	)

	// This hack is to embed taskq logs with logrus
	taskq.SetLogger(stdr.New(stdlibLog.New(log.StandardLogger().WriterLevel(log.WarnLevel), "taskq", 0)))

	log.WithFields(config.SchedulerConfig(cfg.Pull.Metrics).Log()).Info("pull metrics")
	log.WithFields(config.SchedulerConfig(cfg.GarbageCollect.Metrics).Log()).Info("garbage collect metrics")

	return
}

func exit(exitCode int, err error) cli.ExitCoder {
	defer log.WithFields(
		log.Fields{
			"execution-time": time.Since(start),
		},
	).Debug("exited..")

	if err != nil {
		log.WithError(err).Error()
	}

	return cli.Exit("", exitCode)
}

// ExecWrapper gracefully logs and exits our `run` functions.
func ExecWrapper(f func(ctx *cli.Context) (int, error)) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return exit(f(ctx))
	}
}

// configCliOverrides when using the CLI args win over config.
func configCliOverrides(ctx *cli.Context, cfg *config.Config) {
	if cfg.Server.Webhook.Enabled {
		if ctx.String("webhook-secret-token") != "" {
			cfg.Server.Webhook.SecretToken = ctx.String("webhook-secret-token")
		}
	}

	if ctx.String("redis-url") != "" {
		cfg.Redis.URL = ctx.String("redis-url")
	}
}

func assertStringVariableDefined(ctx *cli.Context, k string) {
	if len(ctx.String(k)) == 0 {
		_ = cli.ShowAppHelp(ctx)

		log.Errorf("'--%s' must be set!", k)
		os.Exit(2)
	}
}
