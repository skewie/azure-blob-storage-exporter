package metrics

import "github.com/prometheus/client_golang/prometheus"

const collectorNamespace = "azure_blob_storage_exporter"

type AzureCollector struct {
	blobSizeVec      *prometheus.GaugeVec
	blobCreatedAtVec *prometheus.GaugeVec
}

func NewAzureCollector() *AzureCollector {
	return &AzureCollector{
		blobSizeVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: collectorNamespace,
				Name:      "blob_size",
				Help:      "size of the blob, labeled by name",
			}, []string{"name"},
		),
		blobCreatedAtVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: collectorNamespace,
				Name:      "blob_created_at",
				Help:      "unix time of the blob creation time, labeled by name",
			}, []string{"name"},
		),
	}
}

func (collector AzureCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.blobSizeVec.Describe(ch)
	collector.blobCreatedAtVec.Describe(ch)
}

func (collector AzureCollector) Collect(ch chan<- prometheus.Metric) {
	collector.blobSizeVec.Collect(ch)
	collector.blobCreatedAtVec.Collect(ch)
}

// TrackBlobSize tracks the specific size of a blob object
func (collector *AzureCollector) TrackBlobSize(name string, size float64) {
	collector.blobSizeVec.WithLabelValues(name).Set(size)
}

// TrackBlobCreateTime tracks the specific created at time of a blob object
func (collector *AzureCollector) TrackBlobCreateTime(name string, createdAt float64) {
	collector.blobCreatedAtVec.WithLabelValues(name).Set(createdAt)
}
