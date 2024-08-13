package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Appointment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	VehicleBrand    string             `bson:"vehicle_brand" json:"vehicle_brand"`
	VehicleModel    string             `bson:"vehicle_model" json:"vehicle_model"`
	ServiceType     string             `bson:"servicetype" json:"service_type"`
	AppointmentDate string             `bson:"appointmentdate" json:"appointment_date"`
	PickupAddress   string             `bson:"pickupaddress" json:"pickup_address"`
	DropoffAddress  string             `bson:"dropoffaddress" json:"dropoff_address"`
	Status          string             `bson:"status" json:"status"`
	CreatedAt       time.Time          `bson:"createdat" json:"created_at"`
	UserName        string             `bson:"username" json:"user_name"`
}
