package main

import (
	"context"
	"encoding/json"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/emailsending"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/providerwebhook"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/store"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

var handlerInstance *providerwebhook.Handler
var authHeaderName string
var authSecret string

func init() {
	ctx := context.Background()

	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatalf("Provider webhook initialization failed; details omitted")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("Provider webhook initialization failed; details omitted")
	}

	db := dynamodb.NewFromConfig(cfg)
	sfnClient := sfn.NewFromConfig(cfg)

	authHeaderName = os.Getenv("PROVIDER_WEBHOOK_AUTH_HEADER_NAME")
	authSecretArn := os.Getenv("PROVIDER_WEBHOOK_AUTH_SECRET_ARN")
	if authHeaderName == "" || authSecretArn == "" {
		log.Fatal("Provider webhook initialization failed; details omitted")
	}

	cache, err := secretcache.New()
	if err != nil {
		log.Fatalf("Provider webhook initialization failed; details omitted")
	}
	authSecret, err = cache.GetSecretString(authSecretArn)
	if err != nil {
		log.Fatalf("Provider webhook initialization failed; details omitted")
	}

	provisioningTableName := os.Getenv("PROVISIONING_REQUESTS_TABLE")
	if provisioningTableName == "" {
		log.Fatal("Provider webhook initialization failed; details omitted")
	}

	unprovisioningTableName := os.Getenv("UNPROVISIONING_REQUESTS_TABLE")
	if unprovisioningTableName == "" {
		log.Fatal("Provider webhook initialization failed; details omitted")
	}

	emailCommandTableName := os.Getenv("EMAIL_COMMANDS_TABLE")
	if emailCommandTableName == "" {
		log.Fatal("Provider webhook initialization failed; details omitted")
	}

	provisioningStore := &store.ProvisioningStore{
		TableName: provisioningTableName,
		Client:    db,
	}

	unprovisioningStore := &store.UnprovisioningStore{
		TableName: unprovisioningTableName,
		Client:    db,
	}

	emailCommandStore := &emailsending.EmailCommandStore{
		TableName: emailCommandTableName,
		Client:    db,
	}

	emailStateMachineArn := os.Getenv("EMAIL_STATE_MACHINE_ARN")
	if emailStateMachineArn == "" {
		log.Fatal("Provider webhook initialization failed; details omitted")
	}

	emailStateMachine := &emailsending.StateMachine{
		Client:          sfnClient,
		StateMachineArn: emailStateMachineArn,
	}

	handlerInstance = providerwebhook.NewHandler(
		provisioningStore,
		unprovisioningStore,
		emailCommandStore,
		emailStateMachine,
	)
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
	authHeader := getHeaderValue(&event.Headers, authHeaderName)
	if !providerwebhook.ValidSharedSecret(authSecret, authHeader) {
		log.Printf("Provider webhook status update")
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       `{"error": "unauthorized"}`,
		}, nil
	}

	var payload providerwebhook.WebhookPayload
	if err := json.Unmarshal([]byte(event.Body), &payload); err != nil {
		log.Printf("Provider webhook operation failed; details omitted")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "invalid payload"}`,
		}, nil
	}

	if err := handlerInstance.Handle(ctx, payload); err != nil {
		log.Printf("Provider webhook operation failed; details omitted")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"status": "accepted"}`,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
