package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	httpClient = &http.Client{
		Timeout: 3 * time.Second,
	}
	listenAddress = flag.String(
		"listen", ":9104",
		"Address to listen on for web interface and telemetry.",
	)
	sensuAPI = flag.String(
		"api", "http://10.140.131.43:4567/results",
		"Address to Sensu API.",
	)
	metricsExported map[string]prometheus.Gauge = make(map[string]prometheus.Gauge)
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

func main() {

	go serveMetrics()

	for {
		err := getSensuResults(*sensuAPI)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(3 * time.Second)
	}
}

func serveMetrics() {
	metricPath := "/metrics"
	http.Handle(metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(metricPath))
	})
	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func getSensuResults(url string) error {
	log.Infoln("getSensuResults", url)
	results := []SensuCheckResult{}
	err := getJson(url, &results)
	if err != nil {
		return err
	}
	for i, result := range results {
		log.Infoln("...", fmt.Sprintf("%d, %v, %v", i, result.Check.Name, result.Check.Status))
		checkName := strings.Replace(result.Check.Name, "-", "_", -1)
		elem, ok := metricsExported[checkName]
		if ok {
			elem.Set(float64(result.Check.Status))
		} else {
			gauge := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: checkName,
				Help: "Sensu Check Status",
			})
			gauge.Set(float64(result.Check.Status))
			prometheus.MustRegister(gauge)
			metricsExported[checkName] = gauge
		}
	}
	return nil
}

func getJson(url string, obj interface{}) error {
	log.Infoln("getJson", url)
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}
