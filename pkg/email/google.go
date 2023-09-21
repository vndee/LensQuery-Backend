package email

import (
	"bytes"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strconv"

	"github.com/vndee/lensquery-backend/pkg/model"
	"github.com/vndee/lensquery-backend/pkg/templates"
)

const (
	GoogleSMTPServer = "smtp.gmail.com"
	GoogleSMTPPort   = 465
)

var (
	GoogleEmail    = os.Getenv("GOOGLE_EMAIL")
	GooglePassword = os.Getenv("GOOGLE_PASSWORD")
)

func Send(eventType string, recipient string, data model.EmailData) error {
	var tmpl template.Template
	switch eventType {
	case "INIT_PURCHASE":
		tmpl = *templates.EmailTemplates.InitialPurchase
	case "RENEWAL":
		tmpl = *templates.EmailTemplates.Renewal
	case "CANCELATION":
		tmpl = *templates.EmailTemplates.Cancelation
	case "EXPIRATION":
		tmpl = *templates.EmailTemplates.Expiration
	default:
		return os.ErrInvalid
	}

	var emailBody bytes.Buffer
	err := (&tmpl).Execute(&emailBody, data)
	if err != nil {
		log.Println("[Err]", err)
		return err
	}

	auth := smtp.PlainAuth("", GoogleEmail, GooglePassword, GoogleSMTPServer)
	subject := "Subject: Your Subject Here!\r\n"
	from := "From: " + GoogleEmail + "\r\n"
	toHeader := "To: " + recipient + "\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	headers := subject + from + toHeader + mime
	msg := []byte(headers + "\r\n" + emailBody.String())

	to := []string{recipient}
	err = smtp.SendMail(GoogleSMTPServer+":"+strconv.Itoa(GoogleSMTPPort), auth, GoogleEmail, to, msg)
	if err != nil {
		return err
	}

	return nil
}
