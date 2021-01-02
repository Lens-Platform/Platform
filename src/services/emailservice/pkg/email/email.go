package email

import (
	"encoding/json"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"github.com/BlackspaceInc/Backend/common/schema/models/schema/proto/contracts"
	"github.com/BlackspaceInc/email-service/pkg/counters"
)

type MailClient struct {
	Mailer   *Mailer
	Client   *sendgrid.Client
	Logger   *zap.Logger
	Telemery *counters.Telemetry
}

func NewEmail(from, subject, body, to, firstName, lastName string, emailType contracts.EmailType) *contracts.EmailContract {
	return &contracts.EmailContract{
		Sender:               from,
		Target:               to,
		Subject:              subject,
		Message:              body,
		Type:                 emailType,
		Firstname:            firstName,
		Lastname:             lastName,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
}

func NewMailClient(apiKey string, logger *zap.Logger, telemetry *counters.Telemetry) *MailClient {
	mailer := NewMailer()
	client := sendgrid.NewSendClient(apiKey)
	return &MailClient{
		Mailer:   mailer,
		Client:   client,
		Logger:   logger,
		Telemery: telemetry,
	}
}

func (client *MailClient) Callback() func(amqp.Delivery) error {
	handler := func(delivery amqp.Delivery) error {
		client.Logger.Info("received message", zap.Any("Message", delivery.Body))

		telemetry := client.Telemery.ServiceCounters
		// increment the number of received emails counter
		telemetry.NumEmailsReceived.Inc()

		email := new(contracts.EmailContract)
		err := json.Unmarshal(delivery.Body, email)
		if err != nil {
			telemetry.FailedEmailUnMarshallingEvents.Inc()
			client.Logger.Error(err.Error())
			return err
		}

		client.Logger.Info("data:", zap.Any("email", email))
		resp, err := client.SendEmail(*email)
		if err != nil {
			telemetry.FailedSendgridRequestCount.Inc()
			client.Logger.Error(err.Error())
			return err
		}

		// increment success
		telemetry.SuccessfulSendgridRequestCount.Inc()
		telemetry.NumEmailsSent.Inc()

		client.Logger.Info("email successfully sent", zap.Any("response", resp))

		if err = delivery.Ack(false); err != nil {
			telemetry.NumMessageNacks.Inc()
			client.Logger.Error(err.Error())
			return err
		}

		telemetry.NumMessageAcks.Inc()
		return nil
	}

	return handler
}

func (client *MailClient) SendEmail(email contracts.EmailContract) (string, error) {
	to := mail.NewEmail("", email.Target)
	from := mail.NewEmail("CUBE Platform", email.Sender)
	subject := email.Subject

	switch email.Type {
	case contracts.EmailType_welcome:
		return client.sendWelcomeEmail(email, from, subject, to)
	case contracts.EmailType_invite_code:
		return client.sendInviteEmail(email, from, subject, to)
	case contracts.EmailType_reset_email:
		return client.sendResetEmail(email, from, subject, to)
	default:
		return client.sendWelcomeEmail(email, from, subject, to)
	}
}

func (client *MailClient) sendWelcomeEmail(email contracts.EmailContract, from *mail.Email, subject string, to *mail.Email) (string, error) {
	plainTextContent, err := client.Mailer.WelcomeEmailPlaintext(email.Firstname, email.Lastname)
	if err != nil {
		return "", err
	}

	htmlContent, err := client.Mailer.WelcomeEmailHTML(email.Firstname, email.Lastname)
	if err != nil {
		return "", err
	}

	return client.send(from, subject, to, plainTextContent, htmlContent)
}

func (client *MailClient) sendInviteEmail(email contracts.EmailContract, from *mail.Email, subject string, to *mail.Email) (string, error) {
	plainTextContent, err := client.Mailer.InviteEmailPlaintext(email.Firstname, email.Lastname)
	if err != nil {
		return "", err
	}

	htmlContent, err := client.Mailer.InviteEmailHTML(email.Firstname, email.Lastname)
	if err != nil {
		return "", err
	}

	return client.send(from, subject, to, plainTextContent, htmlContent)
}

func (client *MailClient) sendResetEmail(email contracts.EmailContract, from *mail.Email, subject string, to *mail.Email) (string, error) {
	plainTextContent, err := client.Mailer.ResetEmailPlaintext(email.Firstname, email.Lastname)
	if err != nil {
		return "", err
	}

	htmlContent, err := client.Mailer.ResetEmailHTML(email.Firstname, email.Lastname)
	if err != nil {
		return "", err
	}

	return client.send(from, subject, to, plainTextContent, htmlContent)
}

func (client *MailClient) send(from *mail.Email, subject string, to *mail.Email, plainTextContent string, htmlContent string) (string, error) {
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	start := time.Now()
	response, err := client.Client.Send(message)
	duration := time.Since(start)
	if err != nil {
		return "", err
	}

	client.Telemery.ServiceCounters.SendgridEmailLatency.WithLabelValues("duration").Observe(duration.Seconds())
	client.Logger.Info("Response", zap.Any("Body", response.Body), zap.Any("StatusCode", response.StatusCode), zap.Any("Header", response.Headers))

	return response.Body, nil
}

/*
{
   "target":"yoanyombapro@gmail.com",
   "sender":"yoanyombapro@gmail.com",
   "subject":"hey man how are you",
   "message":"welcome to cube",
   "type": 0
}
*/
