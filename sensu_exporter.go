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
		// exporter port list:
		// https://github.com/prometheus/prometheus/wiki/Default-port-allocations
		"listen", ":9251",
		"Address to listen on for serving Prometheus Metrics.",
	)
	sleepTime = flag.Int("sleep", 10, "sleep seconds between cycles")
	sensuAPI = flag.String(
		"api", "http://localhost:4567",
		"Address to Sensu API.",
	)
	checkStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_check_status",
			Help: "Sensu Check Status(1:Up, 0:Down)",
		},
		[]string{"client", "check_name"},
	)
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

	flag.Parse()
	go serveMetrics()

	for {
		err := getSensuResults(*sensuAPI)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(*sleepTime) * time.Second)
	}
}

func serveMetrics() {
	prometheus.MustRegister(checkStatus)
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
	err := getJson(url+"/results", &results)
	if err != nil {
		return err
	}
	for i, result := range results {
		log.Infoln("...", fmt.Sprintf("%d, %v, %v", i, result.Check.Name, result.Check.Status))
		// in Sensu, 0 means OK
		// in Prometheus, 1 means OK
		status := 0.0
		if result.Check.Status == 0 {
			status = 1.0
		} else {
			status = 0.0
		}
		checkStatus.WithLabelValues(
			result.Client,
			result.Check.Name).Set(status)
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
