package main

import (
        "net/http"
        "time"
        // "bytes"
        "fmt"
        "log"
        "io/ioutil"
        "encoding/json"

        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
        "github.com/prometheus/client_golang/prometheus/promhttp"
)


// [{"date":1592214233.021187,"log":"[2020/06/15 09:43:53] [ warn] [engine] failed to flush chunk '16-1592214228.88222714.flb', retry in 6 seconds: task_id=3, input=alertrule_match > output=http.2\n","time":"2020-06-15T09:43:53.021187056Z","kubernetes":{"pod_name":"fluent-bit-cvxgx","namespace_name":"lma","host":"com1","container_name":"fluent-bit","container_image":"registry.cicd.stg.taco/fluent-bit-skt:1.4.5"},"cluster":"siim-dev"},{"date":1592214236.012278,"log":"[2020/06/15 09:43:56] [ warn] [engine] chunk '16-1592214223.161445788.flb' cannot be retried: task_id=4, input=alertrule_match > output=http.2\n","time":"2020-06-15T09:43:56.012277869Z","kubernetes":{"pod_name":"fluent-bit-cvxgx","namespace_name":"lma","host":"com1","container_name":"fluent-bit","container_image":"registry.cicd.stg.taco/fluent-bit-skt:1.4.5"},"cluster":"siim-dev"}]
type fblog struct {
	// IP address or hostname of the target HTTP Server
    Date float64  `json:"date,omitempty"`
    Log string `json:"log,omitempty"`
    Time string `json:"time,omitempty"`
    // Kubernetes kubernetes `json:"kubernetes,omitempty"`
    Cluster string `json:"cluster,omitempty"`

    PodName string `json:"pod_name,omitempty"`
    NamespaceName string `json:"namespace_name,omitempty"`
    Host string `json:"host,omitempty"`
    ContainerName string `json:"container_name,omitempty"`
    ContainerImage string `json:"container_image,omitempty"`
}

type kubernetes struct {
    PodName string `json:"pod_name,omitempty"`
    NamespaceName string `json:"namespace_name,omitempty"`
    Host string `json:"host,omitempty"`
    ContainerName string `json:"container_name,omitempty"`
    ContainerImage string `json:"container_image,omitempty"`
}

func recordMetrics() {
    go func() {
        for {
            opsProcessed.Inc()
            time.Sleep(2 * time.Second)
        }
    }()
}

var (
    opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
            Name: "myapp_processed_ops_total",
            Help: "The total number of processed events",
    })
)

func collect(w http.ResponseWriter, req *http.Request) {
    for name, headers := range req.Header {
        for _, h := range headers {
            fmt.Printf("%v: %v\n", name, h)
        }
    }

    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        http.Error(w, "can't read body", http.StatusBadRequest)
        return
    }

    var fblogs []fblog
    json.Unmarshal(body, &fblogs)
    for idx, fbl := range fblogs {
        // fmt.Println(idx, "=", string(fbl))
        fmt.Println(idx, "=", fbl.PodName, "=",fbl.Host)

        // opsQueued.WithLabelValues(fbl.PodName, fbl.Host, fbl.Log).Set(1)
        opsQueued.WithLabelValues(fbl.PodName, fbl.Log).Set(1)
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
    recordMetrics()


    opsQueued = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "taco",
            Subsystem: "logging_system",
            Name:      "alert",
            Help:      "Number of blob storage operations waiting to be processed, partitioned by user and type.",
        },
        []string{
            // Which user has requested the operation?
            "podname",
            // Of what type is the operation?
            "log",
        },
    )
    prometheus.MustRegister(opsQueued)
    
    // Increase a value using compact (but order-sensitive!) WithLabelValues().
    opsQueued.WithLabelValues("bob", "put").Add(4)

    metricsPath:="/metrics"

    http.Handle(metricsPath, metrics(promhttp.Handler()))
    http.HandleFunc("/test", collect)
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