package email

import (
	"fmt"

	"github.com/noorbala7418/grafana-oncall-notifier/internal/models"
	"github.com/sirupsen/logrus"
	gomail "gopkg.in/mail.v2"
)

/// SendMail sends email using defined credentials.
func SendMail(email models.Email, cred models.EmailCredential) error {
	message := gomail.NewMessage()

	message.SetHeader("From", email.Sender)
	message.SetHeader("To", email.Receiver)
	message.SetHeader("Subject", email.Subject)
	message.SetBody("text/plain", email.Body)

	dialer := gomail.NewDialer(cred.Host, cred.Port, cred.Username, cred.Password)

	if err := dialer.DialAndSend(message); err != nil {
		logrus.Error("function SendMail. Error in send message. err: ", err)
		return fmt.Errorf("function SendMail. Error in send message. err: %s", err)
	}
	return nil
}
