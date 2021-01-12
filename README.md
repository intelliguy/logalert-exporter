# logalert-exporter
logalert-exporter is a simple prometheus exporter to support fluent-bit. (exactly http output of fluent-bit) 
It can also be used others. 
It has a input-interface for a log insertion and a output-interface for a prometheus.

## Input interface
This exporter run a http server using a port(2112).
You can insert log through the url - http://RUNNING_EXPORTER_SERVER:2112/log.

Format for this should be json with keys - log, cluster, pod, namespace, hosts, level, container....

## Output interface (exporter)
Prometheus can scrap information from the url - http://RUNNING_EXPORTER_SERVER:2112/metrics.
This interface exposes the taco_logging_system_alert metric and other gerneral metrics from prometheus exporters.
taco_logging_system_alert metric always have value=1.
Labels are used for identifing.
Labels have information like below.

- log
- cluster
- pod
- namespace
- container
- level

The below are a example which scraped on a prometheus server.
```
taco_logging_system_alert{cluster="siim-dev",instance="10.233.52.154:2112",job="fluentbit-operator-exporter",kubernetes_name="fluentbit-operator-exporter",kubernetes_namespace="lma",level="critical",log="E0617 11:04:48.251920 1 status.go:71] apiserver received an error that is not an metav1.Status: &errors.errorString{s:"http2: stream closed"} "}
```

## Best Practice
TACO (Kubernetes Distro. from SKT) uses it to implement a alerming from petterns on log.
