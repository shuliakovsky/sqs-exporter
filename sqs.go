package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSAPI interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
}

func receiveMessages(ctx context.Context, client SQSAPI, queueURL string) ([]types.Message, error) {
	output, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     10,
	})
	if err != nil {
		return nil, err
	}
	return output.Messages, nil
}
