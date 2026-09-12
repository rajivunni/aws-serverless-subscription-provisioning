package emailsending

import (
	"context"
	"fmt"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/email"
	"html"
	"log"
)

type Handler struct {
	store  *EmailCommandStore
	sender *email.Sender
}

func NewHandler(store *EmailCommandStore, sender *email.Sender) *Handler {
	return &Handler{
		store:  store,
		sender: sender,
	}
}

func (h *Handler) Handle(ctx context.Context) error {
	log.Printf("Email delivery status update")

	commands, err := h.store.GetPendingCommands(ctx)
	if err != nil {
		log.Printf("Email delivery operation failed; details omitted")
		return err
	}

	log.Printf("Email delivery status update")

	for _, command := range commands {
		log.Printf("Email delivery status update")

		var sendErr error
		switch command.EmailType {
		case ProvisioningComplete:
			sendErr = h.sendProvisioningCompleteEmail(ctx, &command)
		case UnprovisioningComplete:
			sendErr = h.sendUnprovisioningCompleteEmail(ctx, &command)
		default:
			sendErr = fmt.Errorf("unknown email type: %s", command.EmailType)
		}

		if sendErr != nil {
			log.Printf("Email delivery operation failed; details omitted")
			_ = h.store.UpdateStatus(ctx, command.Id, "failed")
			continue
		}

		err = h.store.UpdateStatus(ctx, command.Id, "sent")
		if err != nil {
			log.Printf("Email delivery operation failed; details omitted")
			continue
		}

		log.Printf("Email delivery status update")
	}

	log.Printf("Email delivery status update")
	return nil
}

func (h *Handler) sendProvisioningCompleteEmail(ctx context.Context, command *EmailCommand) error {
	subject := "Example subscription activation update"
	body := fmt.Sprintf(`<html><body><p>Hello %s,</p><p>An activation update was received for your example subscription.</p><p>This is sample notification content for an integration reference.</p></body></html>`, html.EscapeString(command.RecipientName))

	return h.sender.SendEmail(ctx, email.Data{
		To:      command.RecipientEmail,
		Subject: subject,
		Body:    body,
	})
}

func (h *Handler) sendUnprovisioningCompleteEmail(ctx context.Context, command *EmailCommand) error {
	subject := "Example subscription deactivation update"
	body := fmt.Sprintf(`<html><body><p>Hello %s,</p><p>A deactivation update was received for your example subscription.</p><p>This is sample notification content for an integration reference.</p></body></html>`, html.EscapeString(command.RecipientName))

	return h.sender.SendEmail(ctx, email.Data{
		To:      command.RecipientEmail,
		Subject: subject,
		Body:    body,
	})
}
