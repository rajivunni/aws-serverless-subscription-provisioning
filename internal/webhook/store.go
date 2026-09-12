package webhook

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type EventRecord struct {
	Id           string             `dynamodbav:"eventId"`
	ReceivedAt   int64              `dynamodbav:"receivedAt"`
	Email        string             `dynamodbav:"email"`
	OrderId      string             `dynamodbav:"orderId"`
	Status       SubscriptionStatus `dynamodbav:"status"`
	Msisdn       string             `dynamodbav:"msisdn"`
	Ttl          int64              `dynamodbav:"ttl"`
	CustomerName string             `dynamodbav:"customerName"`
}

type MarkResult int

const (
	UnknownStoreError MarkResult = iota
	DuplicateEvent
	EventStored
)

func (event EventRecord) Key() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(event.Id)
	if err != nil {
		return nil, err
	}
	return map[string]types.AttributeValue{"id": id}, nil
}

type EventStore struct {
	client    *dynamodb.Client
	tableName *string
}

func NewEventStore(cfg aws.Config, tableName string) *EventStore {
	client := dynamodb.NewFromConfig(cfg)
	return &EventStore{
		client:    client,
		tableName: aws.String(tableName),
	}
}

func (st *EventStore) MarkReceived(ctx context.Context, event EventRecord) (MarkResult, error) {
	m, err := attributevalue.MarshalMap(event)
	if err != nil {
		return UnknownStoreError, err
	}
	ptc := dynamodb.PutItemInput{Item: m,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
		TableName:           st.tableName,
	}
	_, err = st.client.PutItem(ctx, &ptc)
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return DuplicateEvent, nil
		}
		return UnknownStoreError, err
	}
	return EventStored, nil
}

func (st *EventStore) FindById(ctx context.Context, id string) (*EventRecord, error) {
	event := &EventRecord{
		Id: id,
	}
	key, err := event.Key()
	if err != nil {
		return nil, err
	}
	cmd := dynamodb.GetItemInput{
		Key:       key,
		TableName: st.tableName,
	}

	out, err := st.client.GetItem(ctx, &cmd)

	if err != nil {
		return nil, err
	}

	if out.Item == nil {
		return nil, nil
	}

	err = attributevalue.UnmarshalMap(out.Item, event)

	return event, err
}
