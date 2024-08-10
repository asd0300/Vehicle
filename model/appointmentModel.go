package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Appointment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	VehicleDetails  string             `json:"vehicle_details"`
	ServiceType     string             `json:"service_type"`
	AppointmentDate time.Time          `json:"appointment_date"`
	PickupAddress   string             `json:"pickup_address"`
	DropoffAddress  string             `json:"dropoff_address"`
	Status          string             `json:"status"`
	CreatedAt       time.Time          `json:"created_at"`
}
