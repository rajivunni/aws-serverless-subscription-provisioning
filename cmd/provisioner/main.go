package main

import (
	"context"
	"errors"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/orchestration"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/provider"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/store"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

var orchestratorInstance *orchestration.Orchestrator

func init() {
	ctx := context.Background()
	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatalf("Provisioner initialization failed; details omitted")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("Provisioner initialization failed; details omitted")
	}
	db := dynamodb.NewFromConfig(cfg)
	chApiUrl := os.Getenv("PROVIDER_API_URL")
	if chApiUrl == "" {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	chAuthUrl := os.Getenv("PROVIDER_AUTH_URL")
	if chAuthUrl == "" {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	chAccountId, err := strconv.Atoi(os.Getenv("PROVIDER_ACCOUNT_ID"))
	if chAccountId == 0 || err != nil {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	chCompanyId, err := strconv.Atoi(os.Getenv("PROVIDER_COMPANY_ID"))
	if chCompanyId == 0 || err != nil {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	holdingTariffId, err := strconv.Atoi(os.Getenv("PROVIDER_HOLDING_TARIFF_ID"))
	if err != nil || holdingTariffId <= 0 {
		log.Fatal("PROVIDER_HOLDING_TARIFF_ID must be a positive integer")
	}
	provisioningOrderId, err := strconv.Atoi(os.Getenv("PROVIDER_PROVISIONING_ORDER_ID"))
	if err != nil || provisioningOrderId <= 0 {
		log.Fatal("PROVIDER_PROVISIONING_ORDER_ID must be a positive integer")
	}
	wfItemsTable := os.Getenv("WORKFLOW_ITEMS_TABLE")
	if wfItemsTable == "" {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	log.Printf("Provisioner status update")
	prItemsTable := os.Getenv("PROVISIONING_REQUESTS_TABLE")
	if prItemsTable == "" {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	unprItemsTable := os.Getenv("UNPROVISIONING_REQUESTS_TABLE")
	if unprItemsTable == "" {
		log.Fatal("Provisioner initialization failed; details omitted")
	}

	chClientSecret := os.Getenv("PROVIDER_API_CLIENT_SECRET_ARN")
	if chClientSecret == "" {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	chClientID := os.Getenv("PROVIDER_API_CLIENT_ID_ARN")
	if chClientID == "" {
		log.Fatal("Provisioner initialization failed; details omitted")
	}

	sm, err := secretcache.New()
	if err != nil {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	cls, err := sm.GetSecretString(chClientSecret)
	if err != nil {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	cli, err := sm.GetSecretString(chClientID)
	if err != nil {
		log.Fatal("Provisioner initialization failed; details omitted")
	}

	chClient, err := provider.NewClient(provider.Options{
		ClientId:            cli,
		ClientSecret:        cls,
		ApiUrl:              chApiUrl,
		AuthUrl:             chAuthUrl,
		AccountId:           chAccountId,
		CompanyId:           chCompanyId,
		ProvisioningOrderId: provisioningOrderId,
	})
	if err != nil {
		log.Fatal("Provisioner initialization failed; details omitted")
	}
	pStore := &store.ProvisioningStore{
		TableName: prItemsTable,
		Client:    db,
	}
	unpStore := &store.UnprovisioningStore{
		TableName: unprItemsTable,
		Client:    db,
	}
	wfStore := orchestration.WorkflowStore{
		Db:        db,
		TableName: aws.String(wfItemsTable),
	}
	ch := provisioning.Provider{
		Api:      chClient,
		PStore:   pStore,
		UnpStore: unpStore,
	}
	orchestratorInstance = &orchestration.Orchestrator{
		HoldingTariffId: holdingTariffId,
		Ch:              &ch,
		Wfs:             &wfStore,
	}
}

func handleRequest(ctx context.Context, event interface{}) error {
	log.Printf("Provisioner status update")
	_, err := orchestratorInstance.Run(ctx)
	if err != nil {
		log.Printf("Provisioner operation failed; details omitted")
		return errors.New("provisioning batch failed; inspect restricted workflow state")
	}
	log.Printf("Provisioner status update")
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
