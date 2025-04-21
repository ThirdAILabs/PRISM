package services

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailNotifier struct {
	client *sendgrid.Client
}

func NewEmailNotifier(apiKey string) *EmailNotifier {
	return &EmailNotifier{
		client: sendgrid.NewSendClient(apiKey),
	}
}

func (e *EmailNotifier) Notify(sender, recipient, subject, plainTextContext, htmlContext string) error {
	from := mail.NewEmail("ThirdAI", sender)
	to := mail.NewEmail("", recipient)

	message := mail.NewSingleEmail(from, subject, to, plainTextContext, htmlContext)

	_, err := e.client.Send(message)
	return err
}
