package email

import (
	"bytes"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gopkg.in/gomail.v2"
	"html/template"
	"net/smtp"
	"os"
	"path"
	"strconv"
)

type SendEmail struct {
	AssetsDirectory string
	SMTP            SmtpConfig
}

type SendEmailInterface interface {
	NewTemplate(placeHolder EmailPlaceHolder, templateHtml string) (*string, error)
	SendEmailNew(template, email, title string) error
}

func NewSendEmail(AssetsDirectory string, SMTP SmtpConfig) *SendEmail {
	return &SendEmail{
		AssetsDirectory: AssetsDirectory,
		SMTP:            SMTP,
	}
}

func (s *SendEmail) NewTemplate(placeHolder EmailPlaceHolder, templateHtml string) (*string, error) {
	var w string

	filePath := path.Join(s.AssetsDirectory, templateHtml)

	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, placeHolder); err != nil {
		return nil, err
	}
	w = tpl.String()
	return &w, nil
}

func (s *SendEmail) SendEmailNew(template, email, title string) error {
	from := mail.NewEmail("Simpplify", s.SMTP.Email)
	subject := title
	to := mail.NewEmail("Destinatário", email)

	plainTextContent := "Versão texto do conteúdo do email."
	htmlContent := template

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	if client == nil {
		return fmt.Errorf("error: failed to instantiate client")
	}

	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("error: failed to send email: %v", err)
	}

	if response == nil {
		return fmt.Errorf("error: email delivery failed, no response received")
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("error: email delivery failed, status code: %d", response.StatusCode)
	}

	return nil
}
