package email

import (
	"context"
	"fmt"

	brevo "github.com/getbrevo/brevo-go/lib"
)

type Config struct {
	FromAddress string
	FromName    string
	APIKey      string
}

type Sender struct {
	config      Config
	brevoClient *brevo.APIClient
}

type Data struct {
	To      string
	Subject string
	Body    string
}

func NewSender(config Config) *Sender {
	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", config.APIKey)

	return &Sender{
		config:      config,
		brevoClient: brevo.NewAPIClient(cfg),
	}
}
func (s *Sender) SendEmail(ctx context.Context, data Data) error {
	sender := brevo.SendSmtpEmailSender{
		Name:  s.config.FromName,
		Email: s.config.FromAddress,
	}

	to := []brevo.SendSmtpEmailTo{
		{Email: data.To},
	}

	email := brevo.SendSmtpEmail{
		Sender:      &sender,
		To:          to,
		Subject:     data.Subject,
		HtmlContent: data.Body,
	}

	_, _, err := s.brevoClient.TransactionalEmailsApi.SendTransacEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to send email via Brevo: %w", err)
	}

	return nil
}
