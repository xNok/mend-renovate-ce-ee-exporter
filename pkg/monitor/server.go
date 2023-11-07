package monitor

import (
	"net"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/config"
	pb "github.com/xnok/mend-renovate-ce-ee-exporter/pkg/monitor/protobuf"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/store"
)

// Server ..
type Server struct {
	pb.UnimplementedMonitorServer

	cfg                      config.Config
	store                    store.Store
	taskSchedulingMonitoring map[schemas.TaskType]*schemas.TaskSchedulingStatus
}

// NewServer ..
func NewServer(
	c config.Config,
	st store.Store,
	tsm map[schemas.TaskType]*schemas.TaskSchedulingStatus,
) (s *Server) {
	s = &Server{
		cfg:                      c,
		store:                    st,
		taskSchedulingMonitoring: tsm,
	}

	return
}

// Serve ..
func (s *Server) Serve(url *url.URL) {
	if url == nil {
		log.Info("internal monitoring listener address not set")

		return
	}

	log.WithFields(
		log.Fields{
			"scheme": url.Scheme,
			"host":   url.Host,
			"path":   url.Path,
		},
	).Info("internal monitoring listener set")

	grpcServer := grpc.NewServer()
	pb.RegisterMonitorServer(grpcServer, s)

	var (
		l   net.Listener
		err error
	)

	switch url.Scheme {
	case "unix":
		unixAddr, err := net.ResolveUnixAddr("unix", url.Path)
		if err != nil {
			log.WithError(err).Fatal()
		}

		if _, err := os.Stat(url.Path); err == nil {
			if err := os.Remove(url.Path); err != nil {
				log.WithError(err).Fatal()
			}
		}

		defer func(path string) {
			if err := os.Remove(path); err != nil {
				log.WithError(err).Fatal()
			}
		}(url.Path)

		if l, err = net.ListenUnix("unix", unixAddr); err != nil {
			log.WithError(err).Fatal()
		}

	default:
		if l, err = net.Listen(
			url.Scheme, url.Host,
		); err != nil {
			log.WithError(err).Fatal()
		}
	}

	defer l.Close()

	if err = grpcServer.Serve(l); err != nil {
		log.WithError(err).Fatal()
	}
}
