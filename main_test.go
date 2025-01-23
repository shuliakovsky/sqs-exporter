package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gvre/awsmock-v2/sqsmock"
)

type mockSQSAPI struct {
	sqsmock.Client
	Messages []types.Message
}

func (m *mockSQSAPI) ReceiveMessage(_ context.Context, _ *sqs.ReceiveMessageInput, _ ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return &sqs.ReceiveMessageOutput{
		Messages: m.Messages,
	}, nil
}

func TestReceiveMessages(t *testing.T) {
	mockClient := &mockSQSAPI{
		Messages: []types.Message{
			{Body: aws.String("Message 1")},
			{Body: aws.String("Message 2")},
		},
	}

	queueURL := "https://example.com/queue"

	msgs, err := receiveMessages(context.Background(), mockClient, queueURL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(msgs) != len(mockClient.Messages) {
		t.Errorf("expected %d messages, got %d", len(mockClient.Messages), len(msgs))
	}
}
func TestListenQueue(t *testing.T) {
	mockClient := &mockSQSAPI{
		Messages: []types.Message{
			{Body: aws.String("Message 1")},
			{Body: aws.String("Message 2")},
		},
	}

	go listenQueue(QueueConfig{URL: "https://example.com/queue"}, "us-west-2")
	msgs, err := receiveMessages(context.Background(), mockClient, "https://example.com/queue")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}
func TestInitMetrics(t *testing.T) {
	initMetrics()

	if prometheus.DefaultRegisterer == nil {
		t.Error("expected DefaultRegisterer to be initialized")
	}
}
func TestUpdateMessageCount(t *testing.T) {
	initMetrics()

	queueURL := "https://example.com/queue"
	updateMessageCount(queueURL, 5)

	metric := sqsMessageCount.With(prometheus.Labels{"queue_url": queueURL})
	if metric == nil {
		t.Fatalf("expected metric to be initialized")
	}

	metricValue := testutil.ToFloat64(metric)
	if metricValue != 5 {
		t.Errorf("expected 5, got %v", metricValue)
	}
}
func TestUpdateMessageAge(t *testing.T) {
	initMetrics()

	queueURL := "https://example.com/queue"
	expectedAge := 15.0
	updateMessageAge(queueURL, expectedAge)

	metric := sqsMessageAge.With(prometheus.Labels{"queue_url": queueURL})
	if metric == nil {
		t.Fatalf("expected metric to be initialized")
	}

	metricValue := testutil.ToFloat64(metric)
	if metricValue != expectedAge {
		t.Errorf("expected %v, got %v", expectedAge, metricValue)
	}
}
