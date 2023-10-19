# Clickhouse Exporter for Prometheus (old clickhouse-server versions)

This is a simple server that periodically scrapes [ClickHouse](https://clickhouse.com/) stats and exports them via HTTP for [Prometheus](https://prometheus.io/)
consumption.

Exporter could used only for old ClickHouse versions, modern versions have embedded prometheus endpoint.
Look details https://clickhouse.com/docs/en/operations/server-configuration-parameters/settings#server_configuration_parameters-prometheus

To run it:

```bash
./clickhouse_exporter [flags]
```

Help on flags:
```bash
./clickhouse_exporter --help
```

## Usage

```bash
# http
./clickhouse-exporter --data-source-dsn=http://localhost:8123?username=xx&password=xx&read_timeout=60s
./clickhouse-exporter -d=http://localhost:8123?username=xx&password=xx&read_timeout=60s
# tcp
./clickhouse-exporter --data-source-dsn=tcp://localhost:9000?username=xx&password=xx&read_timeout=60s
./clickhouse-exporter -d=tcp://localhost:9000?username=xx&password=xx&read_timeout=60s
```


## Build Docker image
```
docker build -t clickhouse-exporter .
```

## Using Docker

```
docker run -d -p 9116:9116 clickhouse-exporter --data-source-dsn=tcp://localhost:9000?&username=xx&password=xx
docker run -d -p 9116:9116 clickhouse-exporter -d=tcp://localhost:9000?&username=xx&password=xx
```

## Sample dashboard
Grafana dashboard could be a start for inspiration https://grafana.com/grafana/dashboards/882-clickhouse
