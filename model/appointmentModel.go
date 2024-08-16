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
	TimeSlot        string             `bson:"timeslot" json:"time_slot"`
	CreatedAt       time.Time          `bson:"createdat" json:"created_at"`
	UserName        string             `bson:"username" json:"user_name"`
	Comments        []Comment          `bson:"comments" json:"comments"`
}

type Comment struct {
	Author  string    `bson:"author" json:"author"`
	Text    string    `bson:"text" json:"text"`
	Created time.Time `bson:"created" json:"created"`
}

type CalendarData struct {
	Day               string `json:"day" bson:"day"`
	LimitAppointment  int    `json:"limit_appointment" bson:"limit_appointment"`
	Reserved          int    `json:"reserved" bson:"reserved"`
	ResdueAppointment int    `json:"resdue_appointment" bson:"resdue_appointment"`
}
