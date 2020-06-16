package main

import (
        "net/http"
        "fmt"
        "log"
        "io/ioutil"
        "encoding/json"

        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
)

type fblog struct {
    Date float64  `json:"date,omitempty"`
    Log string `json:"log,omitempty"`
    Time string `json:"time,omitempty"`
    Cluster string `json:"cluster,omitempty"`

    Pod string `json:"pod_name,omitempty"`
    Namespace string `json:"namespace_name,omitempty"`
    Host string `json:"host,omitempty"`
    Container string `json:"container_name,omitempty"`
    Image string `json:"container_image,omitempty"`
}

func collect(w http.ResponseWriter, req *http.Request) {
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        http.Error(w, "can't read body", http.StatusBadRequest)
        return
    }

    var fblogs []fblog
    json.Unmarshal(body, &fblogs)
    for _, fbl := range fblogs {
        opsQueued.WithLabelValues(fbl.Log, fbl.Cluster, fbl.Pod, fbl.Namespace, fbl.Host, fbl.Container).Set(1)
    }
}

func metrics(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        h.ServeHTTP(w, r) // call original
        opsQueued.Reset()
    })
}

var opsQueued *prometheus.GaugeVec

func main() {
    // recordMetrics()

    opsQueued = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "taco",
            Subsystem: "logging_system",
            Name:      "alert",
            Help:      "Number of blob storage operations waiting to be processed, partitioned by user and type.",
        },
        []string{
            "log",
            "cluster",
            "pod",
            "namespace",
            "host",
            "container",
        },
    )
    prometheus.MustRegister(opsQueued)
    
    metricsPath:="/metrics"

    http.Handle(metricsPath, metrics(promhttp.Handler()))
    http.HandleFunc("/logs", collect)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>LogAlert Exporter</title></head>
             <body>
             <h1>LogAlert Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

    fmt.Printf("Starting server for testing HTTP POST...\n")
    if err := http.ListenAndServe(":2112", nil); err != nil {
        log.Fatal(err)
    }
}