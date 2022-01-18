package provider

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/google/go-querystring/query"
)

type EmailService interface {
	SendActivationEmail(email, name, code string) error
	SendResetPassEmail(email, name, code string) error
}

type emailServices struct {
	address                string
	from                   string
	auth                   smtp.Auth
	verifyEmailTemplate    *template.Template
	verifyUrl              string
	resetPassEmailTemplate *template.Template
	resetPassUrl           string
}

func StaticEmailService(configs *Configs) EmailService {
	verifyEmailTemplate, err := template.ParseFiles("templates/VerifyEmail.html")
	if err != nil {
		panic(err)
	}
	resetPassEmailTemplate, err := template.ParseFiles("templates/ResetPass.html")
	if err != nil {
		panic(err)
	}
	return &emailServices{
		address:                configs.SmtpHost + ":" + configs.SmtpPort,
		from:                   configs.SmtpSender,
		auth:                   smtp.PlainAuth("", configs.SmtpSender, configs.SmtpPassword, configs.SmtpHost),
		verifyEmailTemplate:    verifyEmailTemplate,
		verifyUrl:              configs.VerifyUrl,
		resetPassEmailTemplate: resetPassEmailTemplate,
		resetPassUrl:           configs.ResetPassUrl,
	}
}

func (service *emailServices) SendActivationEmail(email, name, code string) error {
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

func (service *emailServices) SendResetPassEmail(email, name, code string) error {
	// Receiver email address.
	to := []string{
		email,
	}

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: Forgot password \n%s\n\n", mimeHeaders)))

	v, _ := query.Values(struct {
		Code  string `url:"code"`
		Email string `url:"email"`
	}{
		Code:  code,
		Email: email,
	})

	service.resetPassEmailTemplate.Execute(&body, struct {
		Name         string
		ResetPassUrl string
	}{
		Name:         name,
		ResetPassUrl: service.resetPassUrl + "?" + v.Encode(),
	})

	return smtp.SendMail(service.address, service.auth, service.from, to, body.Bytes())
}
