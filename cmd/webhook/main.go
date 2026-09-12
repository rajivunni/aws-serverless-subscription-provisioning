package main

import (
	"context"
	"encoding/base64"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/orchestration"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/webhook"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

var handlerInstance *webhook.Handler

func init() {
	ctx := context.Background()
	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatalf("Subscription webhook initialization failed; details omitted")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("Subscription webhook initialization failed; details omitted")
	}

	tableName := os.Getenv("WEBHOOK_EVENTS_TABLE")
	if tableName == "" {
		log.Fatal("Subscription webhook initialization failed; details omitted")
	}
	log.Printf("Subscription webhook status update")

	secretID := os.Getenv("SUBSCRIPTION_API_SECRET_ARN")
	if secretID == "" {
		log.Fatal("Subscription webhook initialization failed; details omitted")
	}
	wfsTableName := os.Getenv("WORKFLOW_ITEMS_TABLE")
	if wfsTableName == "" {
		log.Fatal("Subscription webhook initialization failed; details omitted")
	}
	wfs := &orchestration.WorkflowStore{
		Db:        dynamodb.NewFromConfig(cfg),
		TableName: aws.String(wfsTableName),
	}

	store := webhook.NewEventStore(cfg, tableName)
	stArn := os.Getenv("PROVISIONING_STATE_MACHINE_ARN")
	if stArn == "" {
		log.Fatal("Subscription webhook initialization failed; details omitted")
	}
	st := &webhook.StateMachine{
		Client:                      sfn.NewFromConfig(cfg),
		ProvisioningStateMachineArn: stArn,
	}

	tariffIdStr := os.Getenv("PROVIDER_TARIFF_ID")
	if tariffIdStr == "" {
		log.Fatal("Subscription webhook initialization failed; details omitted")
	}
	tariffId, err := strconv.Atoi(tariffIdStr)
	if err != nil {
		log.Fatalf("Subscription webhook initialization failed; details omitted")
	}

	boltOnIdStr := os.Getenv("PROVIDER_BOLTON_ID")
	var boltOnId int
	if boltOnIdStr != "" {
		boltOnId, err = strconv.Atoi(boltOnIdStr)
	}
	h, err := webhook.NewHandler(secretID, store, wfs, st, tariffId, boltOnId)
	if err != nil {
		log.Fatalf("Subscription webhook initialization failed; details omitted")
	}

	handlerInstance = h
}

func getHeaderValue(headers *map[string]string, name string) string {
	for k, v := range *headers {
		if strings.ToLower(k) == strings.ToLower(name) {
			return v
		}
	}
	return ""
}

func handleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Subscription webhook status update")
	signature := getHeaderValue(&event.Headers, "X-Platform-Hmac-Sha256")

	payload := []byte(event.Body)
	if event.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(event.Body)
		if err != nil {
			log.Printf("Subscription webhook operation failed; details omitted")
			return events.APIGatewayProxyResponse{StatusCode: 400}, nil
		}
		payload = decoded
	}

	_, err := handlerInstance.Handle(ctx, signature, payload)
	if err != nil {
		log.Printf("Subscription webhook operation failed; details omitted")
		// we do not want to retry on error
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
		}, nil
	}

	log.Printf("Subscription webhook status update")

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
