BINARY_NAME=azure-blob-storage-exporter
VERSION=0.1.0

all: lint build
lint: 
	golint -set_exit_status

build: 
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)

docker-build:
	docker build --build-arg version=$(VERSION) -t $(BINARY_NAME):$(VERSION) .

clean: 
	rm -f $(BINARY_NAME)
