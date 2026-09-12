package emailsending

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

type StateMachine struct {
	Client          *sfn.Client
	StateMachineArn string
}

func (sm *StateMachine) Start(ctx context.Context, executionName string) error {
	_, err := sm.Client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: aws.String(sm.StateMachineArn),
		Input:           aws.String("{}"),
		Name:            aws.String(executionName),
	})

	if err != nil {
		return fmt.Errorf("failed to start email state machine: %w", err)
	}

	log.Printf("Email workflow status update")
	return nil
}
