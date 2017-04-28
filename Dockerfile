FROM  quay.io/prometheus/busybox:latest
MAINTAINER  reachlin@gmail.com

COPY sensu_exporter /bin/sensu_exporter

EXPOSE      9104
ENTRYPOINT  [ "/bin/sensu_exporter" ]
