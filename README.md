# Sensu Exporter [![Build Status](https://travis-ci.org/reachlin/sensu_exporter.svg)][travis]

A Prometheus exporter for Sensu.

This app. will export Sensu check status as Prometheus metrics. So previous Sensu checks can be integrated into Prometheus.

To run it:

```bash
make
./sensu_exporter [flags]
```

## Flags

```
$ ./sensu_exporter --help
Usage of ./sensu_exporter:
  -api string
      Address to Sensu API. (default "http://10.140.131.43:4567")
  -listen string
      Address to listen on for serving Prometheus Metrics. (default ":9104")
  -sleep int
      sleep seconds between cycles (default 10)
```

## Exported Metrics
| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| sensu_check_status | Was the last query of Consul successful | server, client, check_name |


[travis]: https://travis-ci.org/reachlin/sensu_exporter
