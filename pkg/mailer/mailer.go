package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strings"
)

type MailDriver string

const (
	STMP MailDriver = "smtp"
	API  MailDriver = "api"
)

type Config struct {
	Driver     MailDriver
	Host       string
	Port       string
	From       string
	Username   string
	Password   string
	Encryption string
}

type Individual struct {
	Name  string
	Email string
}

type Mailable interface {
	Subject() string
	From() Individual
	To() []Individual
	Content() string
	Data() map[string]any
}

type Mailer struct {
	to  Individual
	cfg Config
}

func New(cfg Config) Mailer {
	return Mailer{cfg: cfg}
}

func (m Mailer) To(email, name string) Mailer {
	m.to = Individual{Email: email, Name: name}
	return m
}

func (m Mailer) Send(mailable Mailable) {
	if m.cfg.Driver == STMP {
		m.sendViaSMTP(mailable)
		return
	}

	m.sendViaAPI(mailable)
}

func (m Mailer) sendViaAPI(mailable Mailable) {

	to := []map[string]string{{"email": m.to.Email, "name": m.to.Name}}
	for _, t := range mailable.To() {
		to = append(to, map[string]string{"email": t.Email, "name": t.Name})
	}

	emailData := map[string]any{
		"from":    map[string]string{"email": m.cfg.From, "name": "Alfredo"},
		"to":      to,
		"subject": mailable.Subject(),
		"html":    m.composeHTML(mailable.Content(), mailable.Data()),
	}
	jsonData, err := json.Marshal(emailData)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", m.cfg.Host, bytes.NewBuffer(jsonData))

	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", m.cfg.Password))
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(body))
}

func (m Mailer) sendViaSMTP(mailable Mailable) {
	from := fmt.Sprintf("%s <%s>", mailable.From().Name, mailable.From().Email)
	to := []string{fmt.Sprintf("%s <%s>", m.to.Name, m.to.Email)}
	for _, t := range mailable.To() {
		to = append(to, fmt.Sprintf("%s <%s>", t.Name, t.Email))
	}

	subject := mailable.Subject()

	msg := []byte("To: " + strings.Join(to, ", ") + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		m.composeHTML(mailable.Content(), mailable.Data()) + "\r\n")

	addr := fmt.Sprintf("%s:%s", m.cfg.Host, m.cfg.Port)

	err := smtp.SendMail(addr, smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host), from, to, msg)
	if err != nil {
		log.Fatalf("failed to send email: %v", err)
	}
}

func (m Mailer) composeHTML(view string, data map[string]any) string {

	tmpl, err := template.ParseFiles(view)
	if err != nil {
		log.Fatal(err.Error())
		return ""
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Fatal(err.Error())
		return ""
	}

	return body.String()
}
