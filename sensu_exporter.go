package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	timeout       = flag.Duration("timeout", 20, "Timeout in seconds for the API request")
	listenAddress = flag.String(
		// exporter port list:
		// https://github.com/prometheus/prometheus/wiki/Default-port-allocations
		"listen", ":9251",
		"Address to listen on for serving Prometheus Metrics.",
	)
	sensuAPI = flag.String(
		"api", "http://localhost:4567",
		"Address to Sensu API.",
	)
	cache = flag.Bool("cache", false, "Enable caching of results.  Reduces scrape time for large results datasets by pulling data between prometheus scrapes instead of blocking.")
)

type SensuCheckResult struct {
	Client string
	Check  SensuCheck
}

type SensuCheck struct {
	Name        string
	Duration    float64
	Executed    int64
	Subscribers []string
	Output      string
	Status      int
	Issued      int64
	Interval    int
}

// BEGIN: Class SensuCollector
type SensuCollector struct {
	apiUrl        string
	mutex         sync.RWMutex
	cli           *http.Client
	CheckStatus   *prometheus.Desc
	enableCache   bool
	cachedResults []SensuCheckResult
}

func (c *SensuCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CheckStatus
}

func (c *SensuCollector) Collect(ch chan<- prometheus.Metric) {
	var results []SensuCheckResult

	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()

	if c.enableCache {
		if c.cachedResults == nil {
			c.cachedResults = c.getCheckResults()
		}
		results = c.cachedResults
		// Update cache results after each call to collect
		go c.updateCache()
	} else {
		results = c.getCheckResults()
	}

	for i, result := range results {
		log.Debugln("...", fmt.Sprintf("%d, %v, %v", i, result.Check.Name, result.Check.Status))
		// in Sensu, 0 means OK
		// in Prometheus, 1 means OK
		status := 0.0
		if result.Check.Status == 0 {
			status = 1.0
		} else {
			status = 0.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.CheckStatus,
			prometheus.GaugeValue,
			status,
			result.Client,
			result.Check.Name,
		)
	}
}

func (c *SensuCollector) updateCache() {
	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()
	c.cachedResults = c.getCheckResults()
}

func (c *SensuCollector) getCheckResults() []SensuCheckResult {
	log.Debugln("Sensu API URL", c.apiUrl)
	results := []SensuCheckResult{}
	err := c.GetJson(c.apiUrl+"/results", &results)
	if err != nil {
		log.Errorln("Query Sensu failed.", fmt.Sprintf("%v", err))
	}
	return results
}

func (c *SensuCollector) GetJson(url string, obj interface{}) error {
	resp, err := c.cli.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

// END: Class SensuCollector

func NewSensuCollector(url string, cli *http.Client, enableCache bool) *SensuCollector {
	return &SensuCollector{
		cli:    cli,
		apiUrl: url,
		CheckStatus: prometheus.NewDesc(
			"sensu_check_status",
			"Sensu Check Status(1:Up, 0:Down)",
			[]string{"client", "check_name"},
			nil,
		),
		enableCache: enableCache,
	}
}

func main() {
	flag.Parse()

	collector := NewSensuCollector(*sensuAPI, &http.Client{
		Timeout: *timeout,
	}, *cache)
	fmt.Println(collector.cli.Timeout)
	prometheus.MustRegister(collector)
	metricPath := "/metrics"
	http.Handle(metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(metricPath))
	})
	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
