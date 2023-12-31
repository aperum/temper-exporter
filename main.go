package main

import (
	"log"

	"encoding/json"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type temperCollector struct {
	internalTemp *prometheus.Desc
	externalTemp *prometheus.Desc
}

type Temp struct {
	InternalTemp float64 `json:"internal temperature"`
	ExternalTemp float64 `json:"external temperature"`
}

type Temps []Temp

func newTemperCollector() *temperCollector {
	return &temperCollector{
		internalTemp: prometheus.NewDesc(
			prometheus.BuildFQName("temper", "", "internal_temp"),
			"Reports the internal temper temperature",
			nil, nil,
		),
		externalTemp: prometheus.NewDesc(
			prometheus.BuildFQName("temper", "", "external_temp"),
			"Reports the external temper temperature",
			nil, nil,
		),
	}
}

func (collector *temperCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.internalTemp
	ch <- collector.externalTemp
}

func (collector *temperCollector) Collect(ch chan<- prometheus.Metric) {
	t := getTemp()

	m1 := prometheus.MustNewConstMetric(collector.internalTemp, prometheus.GaugeValue, t.InternalTemp)
	m2 := prometheus.MustNewConstMetric(collector.externalTemp, prometheus.GaugeValue, t.ExternalTemp)
	m1 = prometheus.NewMetricWithTimestamp(time.Now(), m1)
	m2 = prometheus.NewMetricWithTimestamp(time.Now(), m2)
	ch <- m1
	ch <- m2
}

func getTemp() *Temp {
	var t Temps

	out, err := exec.Command("/usr/local/bin/temper.py", "--json").Output()
	if err != nil {
		log.Printf("Could not open temper script: %s\n", err)

		return &Temp{}
	}

	err = json.Unmarshal(out, &t)
	if err != nil {
		log.Printf("Could not unmarshal output: %s\n", err)

		return &Temp{}
	}

	return &t[0]
}

func main() {
  t := newTemperCollector()
	prometheus.MustRegister(t)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9201", nil))
}
