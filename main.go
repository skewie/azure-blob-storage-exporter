package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ben-st/azure-blob-storage-exporter/pkg/client"
	"github.com/ben-st/azure-blob-storage-exporter/pkg/metrics"
	"github.com/ben-st/azure-blob-storage-exporter/pkg/model"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	buildVersion = "dev"
	port         = flag.String("port", ":8080", "The port to listen for HTTP requests.")
	logLevel     = flag.String("loglevel", "info", "possible loglevels are: info or debug")
	interval     = flag.Int("interval", 60, "interval in which metrics should be updated")
	version      = flag.Bool("version", false, "prints current version")
)

func initialize() (accountName, storageAccountKey, containerName string, err error) {
	flag.Parse()

	if *logLevel != "info" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	accountName = os.Getenv("storageAccountName")
	storageAccountKey = os.Getenv("storageAccountKey")
	containerName = os.Getenv("blobContainerName")

	// since all env variables are mandatory, we need to check if they exist
	if os.Getenv("STORAGE_ACCOUNT_NAME") == "" {
		return "", "", "", fmt.Errorf("error: storage account name is required")
	}

	if os.Getenv("STORAGE_ACCOUNT_KEY") == "" {
		return "", "", "", fmt.Errorf("error: storage account name is required")
	}

	if os.Getenv("BLOB_CONTAINER_NAME") == "" {
		return "", "", "", fmt.Errorf("error: storage account name is required")
	}

	return
}

func main() {
	accountName, storageAccountKey, containerName, err := initialize()
	if err != nil {
		log.Errorf("Couldn't start the application: %v", err)
		os.Exit(1)
	}

	// print the build version, set via ldflags in build step and exit
	if *version {
		fmt.Println(buildVersion)
		os.Exit(0)
	}

	// register prometheus metrics
	azureCollector := metrics.NewAzureCollector()
	prometheus.MustRegister(azureCollector)

	// register a new azure client
	azureClient, err := client.NewAzureClient(accountName, storageAccountKey, containerName)
	if err != nil {
		log.Errorf("could not create new azure client, the error is: %v \n", err)
	}

	// update metrics in a new goroutine, since http server is blocking
	// we could avoid this, using the prometheus collector instead
	go UpdateRoutine(azureClient, azureCollector, 15*time.Second)

	log.Info("server startup suceeded")
	log.Infof("serving metrics on http://localhost%s/metrics", *port)

	// start prometheus handler and serve metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*port, nil))
}

type CloudClient interface {
	GetBlobs() ([]model.BlobMetaInformation, error)
}

type Collector interface {
	TrackBlobSize(name string, size float64)
	TrackBlobCreateTime(name string, createdAt float64)
}

// UpdateMetrics updates prometheus metrics
func UpdateRoutine(client CloudClient, collector Collector, duration time.Duration) {
	for range time.NewTicker(duration).C {
		blobMetaInfos, err := client.GetBlobs()
		if err != nil {
			logrus.WithError(err).Error("Couldn't update metrics")
			continue
		}

		for _, blobMetaInfo := range blobMetaInfos {
			collector.TrackBlobSize(blobMetaInfo.Name, blobMetaInfo.ContentSize)
			collector.TrackBlobCreateTime(blobMetaInfo.Name, blobMetaInfo.CreationTime)
		}
	}
}
