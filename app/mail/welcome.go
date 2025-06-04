package mail

import (
	"github.com/martin3zra/acme/pkg/mailer"
)

type Welcome struct {
}

func (w Welcome) Subject() string { return "Welcome to Acme!" }

func (w Welcome) From() mailer.Individual {
	return mailer.Individual{Name: "Alfredo", Email: "alfredo@example.com"}
}

func (w Welcome) To() []mailer.Individual {
	return []mailer.Individual{
		{Name: "Massiel", Email: "massiel@example.com"},
		{Name: "Natasha", Email: "natasha@example.com"},
		{Name: "Nathalia", Email: "nathalia@example.com"},
	}
}

func (w Welcome) Content() string { return "resources/views/mail/welcome.html" }

func (w Welcome) Data() map[string]any {
	return map[string]any{
		"Subject": "Welcome!",
		"Title":   "Hello from Go 👋",
		"Message": "Thanks for testing email sending with MailHog and HTML templates.",
		"CTAText": "Just Do it",
		"CTAURL":  "http://localhost:8092",
		"Year":    2025,
	}
}
