FROM golang:latest AS builder
ARG version="0.1.0"
WORKDIR /go/src/github.com/ben-st/azure-blob-storage-exporter
COPY . .
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${version}" -o azure-blob-storage-exporter

# Final image.
FROM alpine:3.8
LABEL maintainer="Benjamin Stein <info@diffus.org>"
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/ben-st/azure-blob-storage-exporter/azure-blob-storage-exporter .
EXPOSE 8080
ENTRYPOINT ["/azure-blob-storage-exporter"]
CMD ""