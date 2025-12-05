package mail

import (
	"github.com/martin3zra/acme/pkg/mailer"
)

type Verify struct {
	account   map[string]any
	signedUrl string
}

func NewVerification(account map[string]any, signeUrl string) Verify {
	return Verify{account: account, signedUrl: signeUrl}
}

func (w Verify) Subject() string { return "Verify Account" }

func (w Verify) From() mailer.Individual {
	return mailer.Individual{Name: "Alfredo", Email: "martin3zra@gmail.com"}
}

func (w Verify) To() []mailer.Individual {
	return []mailer.Individual{}
}

func (w Verify) Content() string { return "views/mail/verify.html" }

func (w Verify) Data() map[string]any {

	return map[string]any{
		"Subject": "Verify Account",
		"Title":   "Hola👋, quieres verificar tu cuenta?",
		"Message": "Ahaga click en el boton para verificar su cuenta.",
		"CTAText": "Verificar cuenta",
		"CTAURL":  w.signedUrl,
		"Year":    2025,
	}
}
