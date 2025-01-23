package main

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once            sync.Once
	sqsMessageCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sqs_message_count",
			Help: "Number of messages in the sqs",
		},
		[]string{"queue_url"},
	)
)

func initMetrics() {
	once.Do(func() {
		prometheus.MustRegister(sqsMessageCount)
	})
}

func updateMessageCount(queueURL string, count int) {
	sqsMessageCount.With(prometheus.Labels{"queue_url": queueURL}).Set(float64(count))
}
