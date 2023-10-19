FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

LABEL maintainer="fuxingZhang"

ARG TARGETOS
ARG TARGETARCH

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY clickhouse_exporter.go ./main.go
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o clickhouse_exporter main.go

FROM --platform=$TARGETPLATFORM alpine

WORKDIR /app

COPY --from=builder /app/clickhouse_exporter clickhouse_exporter
# RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ENTRYPOINT ["/app/clickhouse_exporter"]
CMD ["-d=http://localhost:8123"]
EXPOSE 9116
