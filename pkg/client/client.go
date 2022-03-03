package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/am3o/azure-blob-storage-exporter/pkg/model"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type AzureClient struct {
	containerURL azblob.ContainerURL
}

func NewAzureClient(accountName, storageAccountKey, containerName string) (*AzureClient, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, storageAccountKey)
	if err != nil {
		return nil, err
	}
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	storageURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))
	if err != nil {
		return nil, err
	}

	serviceURL := azblob.NewServiceURL(*storageURL, pipeline)

	containerURL := serviceURL.NewContainerURL(containerName)

	return &AzureClient{
		containerURL: containerURL,
	}, nil
}

func (ac *AzureClient) GetBlobs() ([]model.BlobMetaInformation, error) {
	var result []model.BlobMetaInformation

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := ac.containerURL.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return result, err
		}

		// IMPORTANT: ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker
		for _, blobInfo := range listBlob.Segment.BlobItems {
			result = append(result, model.BlobMetaInformation{
				Name:         blobInfo.Name,
				CreationTime: float64(blobInfo.Properties.CreationTime.Unix()),
				ContentSize:  float64(*blobInfo.Properties.ContentLength),
				LastModified: float64(blobInfo.Properties.CreationTime.Unix()),
			})
		}

	}
	return result, nil
}
