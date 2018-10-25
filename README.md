# azure-blob-storage-exporter

prometheus exporter for azure blob storage

lists blobs in given container with their size in bytes and creation time in unix timestamp

## Getting Started

prerequisites:

you have to set the following environment variables:

    export storageAccountName="storage-account-name"
    export storageAccountKey="storage-account-access-key"
    export blobContainerName="container-name"

To run it:

```bash
./azure-blob-storage-exporter
```

Help on flags:

there is only one flag available: port

it will change the default port which is 8080

```bash
./azure-blob-storage-exporter -port="port_number"
```

### Docker

Link to the Docker Hub Repository [azure-blob-storage-exporter](https://hub.docker.com/r/benst/azure-blob-storage-exporter/)

To run the exporter as a Docker container, run:

```bash
docker run -p 8080:8080 -e storageAccountName="storage-account-name" -e storageAccountKey="storage-account-access-key" -e blobContainerName="container-name" benst/azure-blob-storage-exporter:0.1.0
```

### Building

simply build it with:
`make` or
`make build`

or if you want a local docker image:
`make docker-build`

if you don't like Makefiles you can do it yourself with:

```bash
go build -ldflags "-X main.version=0.1.0" -o azure-blob-storage-exporter
```

or with docker:

```bash
docker build --build-arg version=0.1.0 -t azure-blob-storage-exporter:latest .
```

## License

MIT