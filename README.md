# mend-renovate-ce-ee-exporter

Prometheus / OpenMetrics exporter for Mend Renovate insights

`mend-renovate-ce-ee-exporter` allows you to monitor your [Mend Renovate](https://github.com/mend/renovate-ce-ee) with Prometheus or any monitoring solution supporting the OpenMetrics format.

## TL:DR

## Install

```bash
go run github.com/xNok/mend-renovate-ce-ee-exporter/cmd/mend-renovate-ce-ee-exporter@latest
```

## Credit

The structure of this project is taken from (mvisonneau/gitlab-ci-pipelines-exporter)[https://github.com/mvisonneau/gitlab-ci-pipelines-exporter]. Looking into this project was a great opportunity to get started with 
exporter and telemetry setup in Go in general.