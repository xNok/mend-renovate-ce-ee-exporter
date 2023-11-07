package cmd

import (
	"net/url"

	"github.com/urfave/cli/v2"
)

// Global is used for globally shared exporter config.
type Global struct {
	// InternalMonitoringListenerAddress can be used to access
	// some metrics related to the exporter internals
	InternalMonitoringListenerAddress *url.URL
}

func parseGlobalFlags(ctx *cli.Context) (cfg Global, err error) {
	if listenerAddr := ctx.String("internal-monitoring-listener-address"); listenerAddr != "" {
		cfg.InternalMonitoringListenerAddress, err = url.Parse(listenerAddr)
	}

	return
}
