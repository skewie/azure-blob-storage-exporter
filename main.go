package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port = flag.String("port", ":8080", "The port to listen for HTTP requests.")

	blobSizeGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "blob_size",
			Help: "size of the blob, labeled by name",
		},
		[]string{
			"name",
		},
	)

	blobCreatedAtGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "blob_created_at",
			Help: "unix time of the blob creation time, labeled by name",
		},
		[]string{
			"name",
		},
	)
)

func main() {

	storageAccountName := os.Getenv("storageAccountName")
	storageAccountKey := os.Getenv("storageAccountKey")
	blobContainerName := os.Getenv("blobContainerName")

	// since all env variables are mandatory, we need to check if the exist
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

	flag.Parse()

	prometheus.MustRegister(blobCreatedAtGaugeVec)
	prometheus.MustRegister(blobSizeGaugeVec)

	go func(name, key, container string) {

		credential, err := azblob.NewSharedKeyCredential(storageAccountName, storageAccountKey)
		if err != nil {
			log.Fatal(err)
		}

		p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

		u, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", storageAccountName))

		serviceURL := azblob.NewServiceURL(*u, p)

		ctx := context.Background()

		containerURL := serviceURL.NewContainerURL(blobContainerName)
		// default update interval is 60 seconds
		// if needed could be implemented as parameter or env variable
		ticker := time.NewTicker(60 * time.Second)

		// initial metrics update call
		// otherwise metrics are not updated until ticket finishes the first time
		updateMetrics(ctx, containerURL)

		for range ticker.C {

			updateMetrics(ctx, containerURL)

		}

	}(storageAccountName, storageAccountKey, blobContainerName)

	log.Printf("serving metrics on http://localhost%s/metrics", *port)
	// start prometheus handler and serve metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*port, nil))

}

// updateMetrics loops over the given container and updates prometheus metrics for all blobs
func updateMetrics(ctx context.Context, containerURL azblob.ContainerURL) {
	for marker := (azblob.Marker{}); marker.NotDone(); {

		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			log.Fatal(err)
		}

		// IMPORTANT: ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		// set prometheus metrics according to blob properties
		for _, blobInfo := range listBlob.Segment.BlobItems {
			blobName := blobInfo.Name
			blobCreatedAt := float64(blobInfo.Properties.CreationTime.Unix())
			blobSize := float64(*blobInfo.Properties.ContentLength)

			blobSizeGaugeVec.WithLabelValues(blobName).Set(blobSize)
			blobCreatedAtGaugeVec.WithLabelValues(blobName).Set(blobCreatedAt)

		}
	}

}
