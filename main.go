package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configFilePath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg, err := LoadConfig(*configFilePath)
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	initMetrics()

	var wg sync.WaitGroup
	for _, queue := range cfg.Queues {
		wg.Add(1)
		go func(queue QueueConfig) {
			defer wg.Done()
			listenQueue(queue, cfg.AWSRegion)
		}(queue)
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	wg.Wait()
}

func listenQueue(queue QueueConfig, defaultRegion string) {
	region := queue.Region
	if region == "" {
		region = defaultRegion
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load config, %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	for {
		msgs, err := receiveMessages(ctx, sqsClient, queue.URL)
		if err != nil {
			log.Printf("failed to receive messages from %s, %v", queue.URL, err)
			continue
		}

		updateMessageCount(queue.URL, len(msgs))
		time.Sleep(5 * time.Second)
	}
}
