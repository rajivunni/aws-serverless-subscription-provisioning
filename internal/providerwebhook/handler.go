package providerwebhook

import (
	"context"
	"fmt"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/emailsending"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/store"
	"log"
	"strings"

	"github.com/google/uuid"
)

type WebhookPayload struct {
	NotificationType string `json:"NotificationType"`
	Content          struct {
		Id     string `json:"Id"`
		Status string `json:"Status"`
	} `json:"Content"`
}

type Handler struct {
	pStore            *store.ProvisioningStore
	unpStore          *store.UnprovisioningStore
	emailCommandStore *emailsending.EmailCommandStore
	emailStateMachine *emailsending.StateMachine
}

func NewHandler(pStore *store.ProvisioningStore,
	unpStore *store.UnprovisioningStore,
	emailCommandStore *emailsending.EmailCommandStore,
	emailStateMachine *emailsending.StateMachine) *Handler {
	return &Handler{
		pStore:            pStore,
		unpStore:          unpStore,
		emailCommandStore: emailCommandStore,
		emailStateMachine: emailStateMachine,
	}
}

func (h *Handler) Handle(ctx context.Context, payload WebhookPayload) error {
	if strings.ToLower(payload.NotificationType) != "submission" {
		log.Printf("Provider callback status update")
		return nil
	}

	ticketNumber := payload.Content.Id
	status := payload.Content.Status

	log.Printf("Provider callback status update")

	if strings.ToLower(status) != "complete" {
		log.Printf("Provider callback status update")
		return nil
	}

	provisioningRecord, err := h.pStore.GetProvisioning(ctx, ticketNumber)
	if err == nil && provisioningRecord != nil {
		log.Printf("Provider callback status update")
		return h.queueProvisioningCompleteEmail(ctx, provisioningRecord)
	}

	unprovisioningRecord, err := h.unpStore.GetUnprovisioning(ctx, ticketNumber)
	if err == nil && unprovisioningRecord != nil {
		log.Printf("Provider callback status update")
		return h.queueUnprovisioningCompleteEmail(ctx, unprovisioningRecord)
	}

	log.Printf("Provider callback status update")
	return nil
}

func (h *Handler) queueProvisioningCompleteEmail(ctx context.Context, provisioningRecord *store.ProvisioningRecord) error {
	if provisioningRecord.CustomerEmail == "" {
		log.Printf("Provider callback status update")
		return nil
	}

	exists, err := h.emailCommandStore.CommandExistsForTicket(ctx, provisioningRecord.TicketNumber, emailsending.ProvisioningComplete)
	if err != nil {
		return fmt.Errorf("failed to check for existing email command: %w", err)
	}
	if exists {
		log.Printf("Provider callback status update")
		return nil
	}

	customerName := provisioningRecord.CustomerName
	if customerName == "" {
		customerName = "Customer"
	}

	commandId := uuid.New().String()
	command := emailsending.EmailCommand{
		Id:             commandId,
		EmailType:      emailsending.ProvisioningComplete,
		TicketNumber:   provisioningRecord.TicketNumber,
		RecipientName:  customerName,
		RecipientEmail: provisioningRecord.CustomerEmail,
	}

	if err := h.emailCommandStore.InsertCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to store email command: %w", err)
	}

	log.Printf("Provider callback status update")

	return h.emailStateMachine.Start(ctx, fmt.Sprintf("email-%s", commandId))
}

func (h *Handler) queueUnprovisioningCompleteEmail(ctx context.Context, unprovisioningRecord *store.UnprovisioningRecord) error {
	if unprovisioningRecord.CustomerEmail == "" {
		log.Printf("Provider callback status update")
		return nil
	}

	exists, err := h.emailCommandStore.CommandExistsForTicket(ctx, unprovisioningRecord.TicketNumber, emailsending.UnprovisioningComplete)
	if err != nil {
		return fmt.Errorf("failed to check for existing email command: %w", err)
	}
	if exists {
		log.Printf("Provider callback status update")
		return nil
	}

	customerName := unprovisioningRecord.CustomerName
	if customerName == "" {
		customerName = "Customer"
	}

	commandId := uuid.New().String()
	command := emailsending.EmailCommand{
		Id:             commandId,
		EmailType:      emailsending.UnprovisioningComplete,
		TicketNumber:   unprovisioningRecord.TicketNumber,
		RecipientName:  customerName,
		RecipientEmail: unprovisioningRecord.CustomerEmail,
	}

	if err := h.emailCommandStore.InsertCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to store email command: %w", err)
	}

	log.Printf("Provider callback status update")

	return h.emailStateMachine.Start(ctx, fmt.Sprintf("email-%s", commandId))
}
