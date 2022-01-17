package provider

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/google/go-querystring/query"
)

type EmailService interface {
	SendActivationEmail(email string, name string, code string) error
}

type emailServices struct {
	address             string
	from                string
	auth                smtp.Auth
	verifyEmailTemplate *template.Template
	verifyUrl           string
}

func StaticEmailService(configs *Configs) EmailService {
	verifyEmailTemplate, err := template.ParseFiles("/app/templates/VerifyEmail.html")
	if err != nil {
		panic(err)
	}
	return &emailServices{
		address:             configs.SmtpHost + ":" + configs.SmtpPort,
		from:                configs.SmtpSender,
		auth:                smtp.PlainAuth("", configs.SmtpSender, configs.SmtpPassword, configs.SmtpHost),
		verifyEmailTemplate: verifyEmailTemplate,
		verifyUrl:           configs.VerifyUrl,
	}
}

func (service *emailServices) SendActivationEmail(email string, name string, code string) error {
	// Receiver email address.
	to := []string{
		email,
	}

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: Verify your email \n%s\n\n", mimeHeaders)))

	v, _ := query.Values(struct {
		Code  string `url:"code"`
		Email string `url:"email"`
	}{
		Code:  code,
		Email: email,
	})

	service.verifyEmailTemplate.Execute(&body, struct {
		Name            string
		VerificationUrl string
	}{
		Name:            name,
		VerificationUrl: service.verifyUrl + "?" + v.Encode(),
	})

	return smtp.SendMail(service.address, service.auth, service.from, to, body.Bytes())
}
