# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
COPY main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o log-exporter main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/log-exporter .
USER nonroot:nonroot

ENTRYPOINT ["/log-exporter"]