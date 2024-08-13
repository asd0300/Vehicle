package smtp

import (
	"fmt"
	"net/smtp"
	"vehicle/enviroment"
	"vehicle/model"
)

func SendReservationEmail(appointment model.Appointment, user model.User) {
	from := enviroment.SMTPEMAIL
	password := enviroment.SMTPKey
	to := user.Email
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// email content
	subject := "Appointment Confirmation"
	body := fmt.Sprintf("Dear %s,\n\nYour appointment for %s on %s has been confirmed.\n\nThank you!", user.Username, appointment.ServiceType, appointment.AppointmentDate)
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)

	// SMTP auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// send
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		fmt.Println("Failed to send email:", err)
	} else {
		fmt.Println("Email sent successfully to", to)
	}
}
