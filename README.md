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
      Address to Sensu API. (default "http://localhost:4567")
  -listen string
      Address to listen on for serving Prometheus Metrics. (default ":9251")
```

## Exported Metrics
| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| sensu_check_status | Check results in a metric vector, status 1 means OK | client, check_name |


### Docker [![Docker Pulls](https://img.shields.io/docker/pulls/reachlin/sensu_exporter.svg?maxAge=604800)][hub]

To run the sensu exporter as a Docker container, run:

```bash
docker run -p 9251:9251 reachlin/sensu_exporter --api="http://sensu_host:4567"
```

[hub]: https://hub.docker.com/r/reachlin/sensu_exporter/
[travis]: https://travis-ci.org/reachlin/sensu_exporter
