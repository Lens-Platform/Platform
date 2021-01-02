package email

import (
	"fmt"

	"github.com/matcornic/hermes/v2"
)

type Mailer struct {
	Hermes hermes.Hermes
}

func NewMailer() *Mailer {
	return &Mailer{
		Hermes: hermes.Hermes{
			// Optional Theme
			// Theme: new(Default)
			Product: hermes.Product{
				// Appears in header & footer of e-mails
				Name: "CUBE Platform",
				Link: "https://example-hermes.com/",
				// Optional product logo
				Logo:      "http://www.duchess-france.org/wp-content/uploads/2016/01/gopher.png",
				Copyright: "Copyright © 2020 CUBE Platform. All rights reserved.",
			},
		},
	}
}

// welcomeEmailTemplate returns a welcome email template for a given user (first and last name)
func (m *Mailer) welcomeEmailTemplate(firstName, lastName string) *hermes.Email {
	welcomeMsg := fmt.Sprintf("Welcome To The CUBE Family %s ! We are beyond excited to have you onboard. Here at CUBE we hope to provide you with the tools necessary for high level entrepreneural success. We value you as a user as well as your experience on the platform and hope this reflects on your experience. \n\n", firstName)
	// TODO: account activation token should come from the msg posted on the queue
	return &hermes.Email{
		Body: hermes.Body{
			Name:     firstName + " " + lastName,
			Greeting: "CUBE Platform",
			Intros: []string{
				welcomeMsg,
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with CUBE, please click here:",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
			Signature: "yours truly",
		},
	}
}

// WelcomeEmailHTML sends a sample welcome email html template
func (m *Mailer) WelcomeEmailHTML(firstName, lastName string) (string, error) {
	email := m.welcomeEmailTemplate(firstName, lastName)
	return m.Hermes.GenerateHTML(*email)
}

// WelcomeEmailPlaintext sends a sample welcome email html template
func (m *Mailer) WelcomeEmailPlaintext(firstName, lastName string) (string, error) {
	email := m.welcomeEmailTemplate(firstName, lastName)
	return m.Hermes.GeneratePlainText(*email)
}

// ResetEmailTemplate is a template used for password reset requests
func (m *Mailer) resetEmailTemplate(firstName, lastName string) *hermes.Email {
	return &hermes.Email{
		Body: hermes.Body{
			Name: firstName + " " + lastName,
			Intros: []string{
				"You have received this email because a password reset request for you CUBE account was received.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Reset your password",
						Link:  "https://hermes-example.com/reset-password?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

// ResetEmailHTML sends a sample reset email html template
func (m *Mailer) ResetEmailHTML(firstName, lastName string) (string, error) {
	email := m.resetEmailTemplate(firstName, lastName)
	return m.Hermes.GenerateHTML(*email)
}

// ResetEmailPlaintext sends a sample reset plaintext template
func (m *Mailer) ResetEmailPlaintext(firstName, lastName string) (string, error) {
	email := m.resetEmailTemplate(firstName, lastName)
	return m.Hermes.GeneratePlainText(*email)
}

func (m *Mailer) inviteCodeTemplate(firstName, lastName string) *hermes.Email {
	return &hermes.Email{
		Body: hermes.Body{
			Name: firstName + " " + lastName,
			Intros: []string{
				"Welcome to the CUBE family! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Please copy your invite code:",
					// TODO: Make this randomly generated or obtain it from other service through msg posted on the queue
					InviteCode: "123456",
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}
}

// InviteEmailHTML sends a sample invite email html template
func (m *Mailer) InviteEmailHTML(firstName, lastName string) (string, error) {
	email := m.resetEmailTemplate(firstName, lastName)
	return m.Hermes.GenerateHTML(*email)
}

// InviteEmailPlaintext sends a sample invite email plaintext template
func (m *Mailer) InviteEmailPlaintext(firstName, lastName string) (string, error) {
	email := m.resetEmailTemplate(firstName, lastName)
	return m.Hermes.GeneratePlainText(*email)
}
