package smtp

import (
	"fmt"
	"net/smtp"
	"vehicle/enviroment"
	"vehicle/model"

	log "github.com/sirupsen/logrus"
)

func SendReservationEmail(appointment model.Appointment, user model.User) {
	from := enviroment.SMTPEMAIL
	password := enviroment.SMTPKey
	to := user.Email
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// email content
	subject := "Appointment Confirmation"
	body := fmt.Sprintf("Dear %s,\n\nYour appointment for %s on %s has been confirmed.\n\nBrand: %s\tModel: %s\t ServiceType: %s Appointment Date: %s\n\nThank you!", user.Username, appointment.ServiceType, appointment.AppointmentDate, appointment.VehicleBrand, appointment.VehicleModel, appointment.ServiceType, appointment.AppointmentDate)
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)

	// SMTP auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// send
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		log.Warnf("Failed to send email: %v", err)
	} else {
		log.Infof("Email sent successfully to %v", to)
	}
}
