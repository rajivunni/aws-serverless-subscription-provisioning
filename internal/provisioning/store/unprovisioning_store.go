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

type UnprovisioningStore struct {
	Client    *dynamodb.Client
	TableName string
}

type UnprovisioningRecord struct {
	TicketNumber  string `dynamodbav:"ticketNumber"`
	SerialNumber  string `dynamodbav:"serialNumber"`
	Iccid         string `dynamodbav:"iccid"`
	TariffName    string `dynamodbav:"tariffName"`
	Msisdn        string `dynamodbav:"msisdn"`
	Imei          string `dynamodbav:"imei"`
	NetworkId     int    `dynamodbav:"networkId"`
	CustomerName  string `dynamodbav:"customerName,omitempty"`
	CustomerEmail string `dynamodbav:"customerEmail,omitempty"`
	Status        string `dynamodbav:"status"`
	CreatedAt     string `dynamodbav:"createdAt"`
	UpdatedAt     string `dynamodbav:"updatedAt"`
}

func (ur UnprovisioningRecord) Key() map[string]types.AttributeValue {
	key := &types.AttributeValueMemberS{Value: ur.TicketNumber}
	return map[string]types.AttributeValue{"ticketNumber": key}
}

func (us *UnprovisioningStore) InsertUnprovisioning(ctx context.Context, result *provider.UnprovisioningResult, meta *provider.ProvisionOrderMeta) error {
	record := UnprovisioningRecord{
		TicketNumber:  result.TicketNumber,
		SerialNumber:  result.SerialNumber,
		Iccid:         result.Iccid,
		TariffName:    result.TariffName,
		Msisdn:        result.Msisdn,
		Imei:          result.Imei,
		NetworkId:     result.NetworkId,
		Status:        "Created",
		CreatedAt:     time.Now().String(),
		UpdatedAt:     time.Now().String(),
		CustomerName:  meta.CustomerName,
		CustomerEmail: meta.CustomerEmail,
	}
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}
	_, err = us.Client.PutItem(ctx, &dynamodb.PutItemInput{Item: item, TableName: aws.String(us.TableName)})
	return err
}

func (us *UnprovisioningStore) UpdateUnprovisioningStatus(ctx context.Context, ticketNumber, status string) error {
	record := UnprovisioningRecord{
		TicketNumber: ticketNumber,
		Status:       status,
		UpdatedAt:    time.Now().String(),
	}
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}
	_, err = us.Client.PutItem(ctx, &dynamodb.PutItemInput{Item: item,
		ConditionExpression: aws.String("attribute_exists(ticketNumber)"),
		TableName:           aws.String(us.TableName)})
	return err
}

func (us *UnprovisioningStore) GetUnprovisioning(ctx context.Context, ticketNumber string) (*UnprovisioningRecord, error) {
	provisioning := &UnprovisioningRecord{TicketNumber: ticketNumber}
	cmd := dynamodb.GetItemInput{
		TableName: aws.String(us.TableName),
		Key:       provisioning.Key(),
	}
	out, err := us.Client.GetItem(ctx, &cmd)
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, fmt.Errorf("no provisioning found for ticket %s", ticketNumber)
	}
	return provisioning, attributevalue.UnmarshalMap(out.Item, provisioning)
}
