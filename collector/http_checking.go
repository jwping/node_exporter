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
	HttpChecking []string
)

type httpcheckCollector struct {
	logger log.Logger
}

func init() {
	registerCollector("httpcheck", defaultEnabled, NewHttpCheckCollector)
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
		"http_connectivity_detection",
		"Running on each node",
		[]string{"http", "httpStatusCode"}, nil,
	)
	var wg sync.WaitGroup
	wg.Add(len(HttpChecking))
	for _, http := range HttpChecking {
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

func NewHttpCheckCollector(logger log.Logger) (Collector, error) {
	return &httpcheckCollector{logger}, nil
}
