package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ben-st/azure-blob-storage-exporter/pkg/client"
	"github.com/ben-st/azure-blob-storage-exporter/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	buildVersion = "dev"
	port         = flag.String("port", ":8080", "The port to listen for HTTP requests.")
	logLevel     = flag.String("loglevel", "info", "possible loglevels are: info or debug")
	interval     = flag.Int("interval", 60, "interval in which metrics should be updated")
	version      = flag.Bool("version", false, "prints current version")
)

func main() {

	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})

	if *logLevel != "info" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// print the build version, set via ldflags in build step and exit
	if *version {
		fmt.Println(buildVersion)
		os.Exit(0)
	}

	accountName := os.Getenv("storageAccountName")
	storageAccountKey := os.Getenv("storageAccountKey")
	containerName := os.Getenv("blobContainerName")

	// since all env variables are mandatory, we need to check if they exist
	if os.Getenv("storageAccountName") == "" {
		fmt.Println("storageAccountName is not set, exiting")
		os.Exit(1)
	}

	if os.Getenv("storageAccountKey") == "" {
		fmt.Println("storageAccountKey is not set, exiting")
		os.Exit(1)
	}

	if os.Getenv("blobContainerName") == "" {
		fmt.Println("blobContainerName is not set, exiting")
		os.Exit(1)
	}

	// register prometheus metrics
	metrics.New()

	// register a new azure client
	a, err := client.NewAzureClient(accountName, storageAccountKey, containerName)
	if err != nil {
		log.Errorf("could not create new azure client, the error is: %v \n", err)
	}

	// update metrics in a new goroutine, since http server is blocking
	// we could avoid this, using the prometheus collector instead
	go UpdateMetrics(a)

	log.Infof("server startup suceeded")
	log.Infof("serving metrics on http://localhost%s/metrics", *port)

	// start prometheus handler and serve metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*port, nil))

}

// UpdateMetrics updates prometheus metrics
func UpdateMetrics(client client.AzureClient) {

	ticker := time.NewTicker(time.Second * time.Duration(*interval))

	result, err := client.GetBlobs()
	if err != nil {
		log.Errorf("get Blobs failed with: %v \n", err)
	}

	// initial metrics update call
	// otherwise metrics are not updated until ticker finishes the first time
	for _, blob := range result {
		myStruct := &metrics.AzureBlobMetrics{Name: blob.Name, Size: blob.ContentSize, CreationTime: blob.CreationTime}
		metrics.Update(*myStruct)
	}

	// update metrics within ticker interval
	for range ticker.C {
		result, err := client.GetBlobs()
		if err != nil {
			log.Errorf("get Blobs failed with: %v \n", err)
		}
		for _, blob := range result {
			myStruct := &metrics.AzureBlobMetrics{Name: blob.Name, Size: blob.ContentSize, CreationTime: blob.CreationTime}
			metrics.Update(*myStruct)
		}
	}

}
