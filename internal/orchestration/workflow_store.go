package orchestration

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Action string

const (
	Provision   Action = "provision"
	Unprovision Action = "unprovision"
	Suspend     Action = "suspend"
)

type State string

const (
	Pending State = "pending"
	Running State = "running"
	Success State = "success"
	Failure State = "failure"
)

type Source string

const (
	Webhook Source = "webhook"
	Manual  Source = "manual"
)

type WorkflowItem struct {
	Id           string                 `dynamodbav:"id"`
	Action       Action                 `dynamodbav:"action"`
	State        State                  `dynamodbav:"state"`
	Source       Source                 `dynamodbav:"source"`
	Error        string                 `dynamodbav:"error"`
	ExecuteAfter int64                  `dynamodbav:"executeAfter"`
	CreatedAt    int64                  `dynamodbav:"createdAt"`
	UpdatedAt    int64                  `dynamodbav:"updatedAt"`
	Data         map[string]interface{} `dynamodbav:"data"`
}

func (item *WorkflowItem) Key() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: item.Id}}
}

type WorkflowStore struct {
	Db        *dynamodb.Client
	TableName *string
}

func (wfs *WorkflowStore) Add(ctx context.Context, item WorkflowItem) error {
	dbItem, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = wfs.Db.PutItem(ctx, &dynamodb.PutItemInput{Item: dbItem, TableName: wfs.TableName})
	return err
}

func (wfs *WorkflowStore) GetPendingItems(ctx context.Context) ([]WorkflowItem, error) {
	q := dynamodb.QueryInput{TableName: wfs.TableName,
		IndexName:              aws.String("state-index"),
		KeyConditionExpression: aws.String("#s = :state"),
		ExpressionAttributeNames: map[string]string{
			"#s": "state",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":state": &types.AttributeValueMemberS{Value: string(Pending)},
		},
		ScanIndexForward: aws.Bool(true),
	}
	var items []WorkflowItem
	pagination := true
	for pagination {
		resp, err := wfs.Db.Query(ctx, &q)
		if err != nil {
			return nil, err
		}
		pagination = resp.LastEvaluatedKey != nil
		q.ExclusiveStartKey = resp.LastEvaluatedKey
		if resp.Items == nil || len(resp.Items) == 0 {
			break
		}
		var ci []WorkflowItem
		err = attributevalue.UnmarshalListOfMaps(resp.Items, &ci)
		if err != nil {
			return nil, err
		}
		items = append(items, ci...)
	}
	return items, nil
}

func (wfs *WorkflowStore) UpdateState(ctx context.Context, id string, state State, errorMessage string) error {
	in := WorkflowItem{Id: id}
	key := in.Key()
	_, err := wfs.Db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key:              key,
		UpdateExpression: aws.String("set #s = :s, #e = :e"),
		ExpressionAttributeNames: map[string]string{
			"#s": "state",
			"#e": "error",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberS{Value: string(state)},
			":e": &types.AttributeValueMemberS{Value: errorMessage},
		},
		TableName: wfs.TableName})
	return err
}
