package collector

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	portCheckSubsystem = "portcheck"
)

var (
	PortChecking []string
)

type portcheckCollector struct {
	logger log.Logger
}

func init() {
	registerCollector("portcheck", defaultEnabled, NewPortCheckCollector)
}

func probeTCP(ctx context.Context, target string) bool {

	deadline, _ := ctx.Deadline()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp4", target)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Set a deadline to prevent the following code from blocking forever.
	// If a deadline cannot be set, better fail the probe by returning an error
	// now rather than blocking forever.
	if err := conn.SetDeadline(deadline); err != nil {
		return false
	}
	return true
}

func (c *portcheckCollector) Update(ch chan<- prometheus.Metric) error {
	testCTX, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// registry := prometheus.NewRegistry()
	portDesc := prometheus.NewDesc(
		"port_connectivity_detection",
		"Running on each node",
		[]string{"port"}, nil,
	)
	var wg sync.WaitGroup
	wg.Add(len(PortChecking))
	for _, port := range PortChecking {
		go func(port string) {
			if probeTCP(testCTX, port) {
				ch <- prometheus.MustNewConstMetric(
					portDesc,
					prometheus.GaugeValue, 1,
					port,
				)
			} else {
				ch <- prometheus.MustNewConstMetric(
					portDesc,
					prometheus.GaugeValue, 0,
					port,
				)
			}
			wg.Done()
		}(port)
	}
	wg.Wait()

	return nil
}

func NewPortCheckCollector(logger log.Logger) (Collector, error) {
	return &portcheckCollector{logger}, nil
}
