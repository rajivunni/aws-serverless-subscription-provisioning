package main

import (
	"context"
	"errors"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/email"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/emailsending"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	secretsmanager "github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

var handlerInstance *emailsending.Handler

func init() {
	ctx := context.Background()

	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatalf("Email sender initialization failed; details omitted")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("Email sender initialization failed; details omitted")
	}

	db := dynamodb.NewFromConfig(cfg)

	emailCommandTableName := os.Getenv("EMAIL_COMMANDS_TABLE")
	if emailCommandTableName == "" {
		log.Fatal("Email sender initialization failed; details omitted")
	}

	emailCommandStore := &emailsending.EmailCommandStore{
		TableName: emailCommandTableName,
		Client:    db,
	}

	brevoAPIKeyArn := os.Getenv("EMAIL_API_KEY_ARN")
	if brevoAPIKeyArn == "" {
		log.Fatal("Email sender initialization failed; details omitted")
	}

	secretCache, err := secretsmanager.New()
	if err != nil {
		log.Fatalf("Email sender initialization failed; details omitted")
	}

	brevoAPIKey, err := secretCache.GetSecretString(brevoAPIKeyArn)
	if err != nil {
		log.Fatalf("Email sender initialization failed; details omitted")
	}

	fromAddress := os.Getenv("FROM_ADDRESS")
	fromName := os.Getenv("FROM_NAME")

	if fromAddress == "" {
		log.Fatal("Email sender initialization failed; details omitted")
	}

	sender := email.NewSender(email.Config{
		FromAddress: fromAddress,
		FromName:    fromName,
		APIKey:      brevoAPIKey,
	})

	handlerInstance = emailsending.NewHandler(emailCommandStore, sender)
}

func handleRequest(ctx context.Context) error {
	if err := handlerInstance.Handle(ctx); err != nil {
		return errors.New("email batch failed; details withheld")
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
