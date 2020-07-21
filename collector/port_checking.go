// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	PORTChecking []string
)

type portcheckCollector struct {
	logger log.Logger
}

func init() {
	registerCollector(portCheckSubsystem, defaultEnabled, NewPortCheckCollector)
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
		// "port_connectivity_detection",
		prometheus.BuildFQName(namespace, portCheckSubsystem, "Status"),
		"Running on each node",
		[]string{"port"}, nil,
	)
	var wg sync.WaitGroup
	wg.Add(len(PORTChecking))
	for _, port := range PORTChecking {
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
