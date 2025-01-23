// main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Version and CommitHash will be set during the build process
var Version string = "0.0.1" //Default Version value
var CommitHash string = ""

func printVersion() {
	fmt.Printf("sqs-exporter version: %s\n", Version)
	if CommitHash != "" {
		fmt.Printf("commit hash: %s\n", CommitHash)
	}
}
func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  --config <path to config file>     Specify the path to the configuration file.")
	fmt.Println("  --help                             Display this help message.")
	fmt.Println("  --version                          Display the current version of the application.")
}

func printConfiguration(cfg Config) {
	headers := []string{"Queue URL", "Region"}
	maxURLLength := len(headers[0])
	maxRegionLength := len(headers[1])

	for _, queue := range cfg.Queues {
		if len(queue.URL) > maxURLLength {
			maxURLLength = len(queue.URL)
		}
		region := queue.Region
		if len(region) == 0 {
			region = cfg.AWSRegion
		}
		if len(region) > maxRegionLength {
			maxRegionLength = len(region)
		}
	}

	borderLength := maxURLLength + maxRegionLength + 5 // +5 for the spaces and separator
	border := strings.Repeat("═", borderLength)

	fmt.Println(border)
	log.Printf("sqs-autoscaler version: %s commit hash: %s", Version, CommitHash)
	log.Printf("Default AWS Region: %s", cfg.AWSRegion)
	fmt.Println(border)

	format := fmt.Sprintf("%%-%ds | %%-%ds\n", maxURLLength, maxRegionLength)
	fmt.Printf(format, headers[0], headers[1])
	fmt.Println(border)

	for _, queue := range cfg.Queues {
		region := queue.Region
		if len(region) == 0 {
			region = cfg.AWSRegion
		}
		fmt.Printf(format, queue.URL, region)
	}
	fmt.Println(border)
}

func main() {
	configFilePath := flag.String("config", "config.yaml", "Path to the configuration file")
	help := flag.Bool("help", false, "Show help")
	version := flag.Bool("version", false, "Display the current version of the application")
	flag.Parse()

	if *version {
		printVersion()
		return
	}
	if *help {
		printHelp()
		return
	}
	cfg, err := LoadConfig(*configFilePath)
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	printConfiguration(cfg)
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
	address := fmt.Sprintf("%s:%d", cfg.ListenIP, cfg.Port)
	go func() {
		log.Fatal(http.ListenAndServe(address, nil))
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

		for _, msg := range msgs {
			if sentTimestampStr, exists := msg.Attributes["SentTimestamp"]; exists {
				sentTimestamp, err := strconv.ParseInt(sentTimestampStr, 10, 64)
				if err == nil {
					age := time.Since(time.Unix(0, sentTimestamp*int64(time.Millisecond))).Seconds()
					updateMessageAge(queue.URL, age)

				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}
