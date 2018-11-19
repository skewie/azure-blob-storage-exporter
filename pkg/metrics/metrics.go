package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
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

// AzureBlobMetrics struct
type AzureBlobMetrics struct {
	Name         string
	Size         float64
	CreationTime float64
}

// Update updates the given prometheus metrics
func Update(m AzureBlobMetrics) {
	blobSizeGaugeVec.WithLabelValues(m.Name).Set(m.Size)
	blobCreatedAtGaugeVec.WithLabelValues(m.Name).Set(m.CreationTime)
}

// New registers the prometheus metrics
func New() {
	prometheus.MustRegister(blobCreatedAtGaugeVec)
	prometheus.MustRegister(blobSizeGaugeVec)
}
