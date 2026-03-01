package mail

import (
	"github.com/martin3zra/acme/pkg/mailer"
)

type InvoiceMail struct {
	invoice    map[string]any
	attachment []byte
}

func NewInvoiceMail(invoice map[string]any, attachment []byte) InvoiceMail {
	return InvoiceMail{invoice: invoice, attachment: attachment}
}

func (w InvoiceMail) Subject() string { return "A new invoice is available in your account" }

func (w InvoiceMail) To() []mailer.Individual {
	return []mailer.Individual{}
}

func (w InvoiceMail) Content() string { return "resources/views/mail/invoice.html" }

func (w InvoiceMail) Data() map[string]any {

	return map[string]any{
		"Title":   "A new invoice is available in your account",
		"Message": "A new invoice has been generated in your account. The document is attached for your records.",
	}
}

func (w InvoiceMail) Attachments() []mailer.Attachment {
	return []mailer.Attachment{
		{
			Filename: "invoice.pdf",
			Content:  w.attachment,
			MIMEType: "application/pdf",
		},
	}
}
