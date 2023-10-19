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
./clickhouse_exporter -h
```

Credentials(if not default):

via environment variables
```
CLICKHOUSE_USER
CLICKHOUSE_PASSWORD
```

## Build Docker image
```
docker build -t clickhouse-exporter .
```

## Using Docker

```
docker run -d -p 9116:9116 clickhouse-exporter --uri=http://127.0.0.1:8123
docker run -d -p 9116:9116 clickhouse-exporter --uri=tcp://127.0.0.1:9000
```
## Sample dashboard
Grafana dashboard could be a start for inspiration https://grafana.com/grafana/dashboards/882-clickhouse
