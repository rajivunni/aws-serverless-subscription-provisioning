package webhook

import (
	"context"
	"github.com/google/uuid"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

type StateMachine struct {
	Client                      *sfn.Client
	ProvisioningStateMachineArn string
}

func (st *StateMachine) startProvisioner(ctx context.Context, event *EventRecord) error {

	input := &sfn.StartExecutionInput{
		StateMachineArn: &st.ProvisioningStateMachineArn,
		Name:            aws.String("workflow-" + uuid.New().String()),
	}
	_, err := st.Client.StartExecution(ctx, input)
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		return err
	}
	log.Print("Provisioning workflow execution started")
	return nil
}
