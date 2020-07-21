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
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	httpCheckSubsystem = "httpcheck"
)

var (
	HTTPChecking []string
)

type httpcheckCollector struct {
	logger log.Logger
}

func init() {
	registerCollector(httpCheckSubsystem, defaultEnabled, NewHTTPCheckCollector)
}

func probeHTTP(target string) (float64, int) {
	startTime := time.Now().UnixNano()

	client := &http.Client{Timeout: time.Second * 3}
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return float64((time.Now().UnixNano() - startTime) / 1e6), 0
	}

	resp, err := client.Do(req)
	if err != nil {
		return float64((time.Now().UnixNano() - startTime) / 1e6), 0
	}

	defer resp.Body.Close()
	return float64((time.Now().UnixNano() - startTime) / 1e6), resp.StatusCode
}

func (c *httpcheckCollector) Update(ch chan<- prometheus.Metric) error {

	// registry := prometheus.NewRegistry()
	portDesc := prometheus.NewDesc(
		// "http_connectivity_detection",
		prometheus.BuildFQName(namespace, httpCheckSubsystem, "StatusAndDelay"),
		"Running on each node",
		[]string{"http", "httpStatusCode"}, nil,
	)
	var wg sync.WaitGroup
	wg.Add(len(HTTPChecking))
	for _, http := range HTTPChecking {
		go func(http string) {
			runningTime, statusCode := probeHTTP(http)
			ch <- prometheus.MustNewConstMetric(
				portDesc,
				prometheus.GaugeValue, runningTime,
				http, strconv.Itoa(statusCode),
			)
			wg.Done()
		}(http)
	}
	wg.Wait()

	return nil
}

func NewHTTPCheckCollector(logger log.Logger) (Collector, error) {
	return &httpcheckCollector{logger}, nil
}
