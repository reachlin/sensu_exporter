package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	httpClient = &http.Client{
		Timeout: 3*time.Second,
	}
	listenAddress = flag.String(
		"listen", ":9104",
		"Address to listen on for web interface and telemetry.",
	)
	sensuAPI = flag.String(
		"api", "http://10.140.131.43:4567",
		"Address to Sensu API.",
	)
)

type SensuCheckResult struct {
	Client string
}

func main() {
	metricPath := "/metrics"
	http.Handle(metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(metricPath))
	})
	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))

	for {
		log.Fatal(getSensuResults(*sensuAPI))
		time.Sleep(3*time.Second)
	}
}

func getSensuResults(url string) error {
	results := []SensuCheckResult{}
	err := getJson(url, results)
	if err != nil {
		return err
	}
	log.Infoln(fmt.Sprintln(results))
	return nil
}

func getJson(url string, obj interface{}) error {
	resp, err := httpClient.Get(uri)
	if err != nil {
		return error
	}
	defer resp.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}