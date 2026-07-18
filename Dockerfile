FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=2.0.0
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.Date=${DATE}" \
    -o sentinel-agent ./cmd/sentinel-agent

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S sentinel \
    && adduser -S sentinel -G sentinel \
    && mkdir -p /run/sentinel \
    && chown -R sentinel:sentinel /run/sentinel

USER sentinel

COPY --from=builder /app/sentinel-agent /usr/local/bin/sentinel-agent

LABEL org.opencontainers.image.title="Sentinel" \
      org.opencontainers.image.description="AI-powered Context Engine Agent" \
      org.opencontainers.image.source="https://github.com/tu-usuario/sentinel"

EXPOSE 2112

ENTRYPOINT ["/usr/local/bin/sentinel-agent"]
