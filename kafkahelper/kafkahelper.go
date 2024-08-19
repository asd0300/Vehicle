package kafkahelper

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"vehicle/model"

	smtpHelper "vehicle/smtp"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

func SendReservationEmailKafkaProducer(appointment model.Appointment, user model.User) error {
	//set kafka producer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"}, // Kafka broker 地址
		Topic:   "email-topic",              // Kafka 主题
	})

	//create Data
	emailData := map[string]interface{}{
		"appointment": appointment,
		"user":        user,
	}
	messageBytes, err := json.Marshal(emailData)
	if err != nil {
		log.Errorf("Error getting user by ID: %v, fail to Send Mail", err)
		return err
	}

	//send data to kafka
	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(user.Email),
		Value: messageBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to send email message to Kafka: %v", err)
	}

	log.Info("Email message sent to Kafka for ", user.Email)
	return nil
}

func StartKafkaEmailConsumer() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "email-topic",
		GroupID: "email-consumer-group",
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Failed to read message: %v", err)
			continue
		}

		var emailData struct {
			Appointment model.Appointment `json:"appointment"`
			User        model.User        `json:"user"`
		}

		err = json.Unmarshal(msg.Value, &emailData)
		if err != nil {
			log.Printf("Failed to unmarshal email data: %v", err)
			continue
		}

		// use origin sendMail
		smtpHelper.SendReservationEmail(emailData.Appointment, emailData.User)
	}
}

// 檢測
func IsKafkaOn() bool {
	if os.Getenv("USE_KAFKA") == "true" {
		return true
	}
	return false
}
