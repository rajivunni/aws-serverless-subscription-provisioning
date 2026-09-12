package store

import (
	"context"
	"fmt"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/provider"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ProvisioningRecord struct {
	TicketNumber  string `dynamodbav:"ticketNumber"`
	TariffNumber  int    `dynamodbav:"tariffNumber"`
	TariffName    string `dynamodbav:"tariffName"`
	BoltOnNumber  int    `dynamodbav:"boltOnNumber"`
	BoltOnName    string `dynamodbav:"boltOnName"`
	SerialNumber  string `dynamodbav:"serialNumber"`
	Msisdn        string `dynamodbav:"msisdn"`
	Iccid         string `dynamodbav:"iccid"`
	Imei          string `dynamodbav:"imei"`
	NetworkId     int    `dynamodbav:"networkId"`
	CustomerName  string `dynamodbav:"customerName,omitempty"`
	CustomerEmail string `dynamodbav:"customerEmail,omitempty"`
	CreatedAt     string `dynamodbav:"createdAt"`
	UpdatedAt     string `dynamodbav:"updatedAt"`
	Status        string `dynamodbav:"status"`
}

func (r ProvisioningRecord) Key() map[string]types.AttributeValue {
	key := &types.AttributeValueMemberS{Value: r.TicketNumber}
	return map[string]types.AttributeValue{"ticketNumber": key}
}

type ProvisioningStore struct {
	TableName string
	Client    *dynamodb.Client
}

func (store *ProvisioningStore) InsertProvisioning(ctx context.Context, result *provider.ProvisioningResult, meta *provider.ProvisionOrderMeta) error {
	record := ProvisioningRecord{
		TicketNumber:  result.TicketNumber,
		TariffNumber:  result.TariffNumber,
		TariffName:    result.TariffName,
		BoltOnNumber:  result.BoltOnNumber,
		BoltOnName:    result.BoltOnName,
		SerialNumber:  result.SerialNumber,
		Msisdn:        result.Msisdn,
		Iccid:         result.Iccid,
		Imei:          result.Imei,
		NetworkId:     result.NetworkId,
		CreatedAt:     time.Now().String(),
		UpdatedAt:     time.Now().String(),
		Status:        "Created",
		CustomerName:  meta.CustomerName,
		CustomerEmail: meta.CustomerEmail,
	}
	in, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}
	_, err = store.Client.PutItem(ctx, &dynamodb.PutItemInput{Item: in, TableName: aws.String(store.TableName)})
	return err
}

func (store *ProvisioningStore) UpdateProvisioningStatus(ctx context.Context, ticketNumber, status string) error {
	record := ProvisioningRecord{
		TicketNumber: ticketNumber,
		Status:       status,
		UpdatedAt:    time.Now().String(),
	}
	in, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}
	putCmd := dynamodb.PutItemInput{TableName: aws.String(store.TableName),
		ConditionExpression: aws.String("attribute_exists(ticketNumber)"),
		Item:                in,
	}
	_, err = store.Client.PutItem(ctx, &putCmd)
	return err
}

func (store *ProvisioningStore) GetProvisioning(ctx context.Context, ticketNumber string) (*ProvisioningRecord, error) {
	provisioning := &ProvisioningRecord{TicketNumber: ticketNumber}
	cmd := dynamodb.GetItemInput{
		TableName: aws.String(store.TableName),
		Key:       provisioning.Key(),
	}
	out, err := store.Client.GetItem(ctx, &cmd)
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, fmt.Errorf("no provisioning found for ticket %s", ticketNumber)
	}

	return provisioning, attributevalue.UnmarshalMap(out.Item, provisioning)
}
