ARCH=arm64
OUTPUT_DIR=dist

build-WebhookFunction:
	GOOS=linux GOARCH=$(ARCH) go build -trimpath -ldflags="-s -w" -o $(ARTIFACTS_DIR)/bootstrap ./cmd/webhook

build-ProvisionerFunction:
	GOOS=linux GOARCH=$(ARCH) go build -trimpath -ldflags="-s -w" -o $(ARTIFACTS_DIR)/bootstrap ./cmd/provisioner

build-ProviderWebhookFunction:
	GOOS=linux GOARCH=$(ARCH) go build -trimpath -ldflags="-s -w" -o $(ARTIFACTS_DIR)/bootstrap ./cmd/provider-webhook

build-EmailSenderFunction:
	GOOS=linux GOARCH=$(ARCH) go build -trimpath -ldflags="-s -w" -o $(ARTIFACTS_DIR)/bootstrap ./cmd/email-sender

clean:
	rm -rf $(OUTPUT_DIR)
