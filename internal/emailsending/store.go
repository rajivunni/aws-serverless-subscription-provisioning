package emailsending

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type EmailType string

const (
	ProvisioningComplete   EmailType = "provisioning_complete"
	UnprovisioningComplete EmailType = "unprovisioning_complete"
)

type EmailCommand struct {
	Id             string    `dynamodbav:"id"`
	EmailType      EmailType `dynamodbav:"emailType"`
	TicketNumber   string    `dynamodbav:"ticketNumber"`
	RecipientName  string    `dynamodbav:"recipientName"`
	RecipientEmail string    `dynamodbav:"recipientEmail"`
	CreatedAt      string    `dynamodbav:"createdAt"`
	Status         string    `dynamodbav:"status"`
}

func (ec EmailCommand) Key() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: ec.Id}}
}

type EmailCommandStore struct {
	TableName string
	Client    *dynamodb.Client
}

func (store *EmailCommandStore) InsertCommand(ctx context.Context, command EmailCommand) error {
	command.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	command.Status = "pending"

	item, err := attributevalue.MarshalMap(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	_, err = store.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(store.TableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to insert command: %w", err)
	}

	return nil
}

func (store *EmailCommandStore) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := store.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:        aws.String(store.TableName),
		Key:              EmailCommand{Id: id}.Key(),
		UpdateExpression: aws.String("set #s = :s"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberS{Value: status},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (store *EmailCommandStore) GetPendingCommands(ctx context.Context) ([]EmailCommand, error) {
	result, err := store.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(store.TableName),
		FilterExpression: aws.String("#s = :status"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: "pending"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan pending commands: %w", err)
	}

	var commands []EmailCommand
	err = attributevalue.UnmarshalListOfMaps(result.Items, &commands)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal commands: %w", err)
	}

	return commands, nil
}

func (store *EmailCommandStore) CommandExistsForTicket(ctx context.Context, ticketNumber string, emailType EmailType) (bool, error) {
	result, err := store.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(store.TableName),
		FilterExpression: aws.String("ticketNumber = :ticket AND emailType = :type"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ticket": &types.AttributeValueMemberS{Value: ticketNumber},
			":type":   &types.AttributeValueMemberS{Value: string(emailType)},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check for existing command: %w", err)
	}

	return len(result.Items) > 0, nil
}
