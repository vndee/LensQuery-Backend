package email

import (
	"bytes"
	"fmt"
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
	GoogleSMTPPort   = 587
)

var (
	GoogleEmail    = os.Getenv("GOOGLE_EMAIL")
	GooglePassword = os.Getenv("GOOGLE_PASSWORD")
	auth           = smtp.PlainAuth("", GoogleEmail, GooglePassword, GoogleSMTPServer)
)

func Send(eventType string, recipient string, data model.EmailData) error {
	var title string
	var tmpl template.Template

	switch eventType {
	case "INITIAL_PURCHASE":
		title = "Thank you for your purchase!"
		tmpl = *templates.EmailTemplates.InitialPurchase
	case "RENEWAL":
		title = "Your subscription has been renewed!"
		tmpl = *templates.EmailTemplates.Renewal
	case "CANCELLATION":
		title = "Your subscription has been canceled!"
		tmpl = *templates.EmailTemplates.Cancelation
	case "EXPIRATION":
		title = "Your subscription has expired!"
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

	subject := "Subject: " + title + "\r\n"
	from := fmt.Sprintf("From: %s <%s>\r\n", "LensQuery", GoogleEmail)
	toHeader := "To: " + recipient + "\r\n"
	mime := "MIME-version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	headers := subject + from + toHeader + mime
	msg := []byte(headers + "\r\n" + emailBody.String())
	to := []string{recipient}
	err = smtp.SendMail(GoogleSMTPServer+":"+strconv.Itoa(GoogleSMTPPort), auth, GoogleEmail, to, msg)
	if err != nil {
		return err
	}

	return nil
}
