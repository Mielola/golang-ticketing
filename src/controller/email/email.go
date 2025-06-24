package email

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

func SendEmail(c *gin.Context, to string, subject string, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "support@adhivasindo.co.id")
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)

	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		"support@adhivasindo.co.id",
		"ofeo pdkr hrfa xmjq",
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		fmt.Printf("Error sending email: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to send OTP email",
			"error":   err.Error(),
		})
		return err
	}
	return nil
}

func init() {

}
