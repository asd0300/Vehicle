package service

import (
	"context"
	"time"
	mongo "vehicle/database"
	model "vehicle/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetAllAppointment(c *gin.Context) {
	var appointments []model.Appointment

	cursor, err := mongo.AppointmentCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch appointments"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var appointment model.Appointment
		if err := cursor.Decode(&appointment); err != nil {
			c.JSON(500, gin.H{"error": "Failed to decode appointment"})
			return
		}
		appointments = append(appointments, appointment)
	}

	c.JSON(200, appointments)
}

func CreateNewAppointment(c *gin.Context) {
	var appointment model.Appointment
	if err := c.BindJSON(&appointment); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	appointment.ID = primitive.NewObjectID()
	appointment.CreatedAt = time.Now()

	_, err := mongo.AppointmentCollection.InsertOne(context.Background(), appointment)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create appointment"})
		return
	}

	c.JSON(201, appointment)
}

func GetDetailAppointmentById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid appointment ID"})
		return
	}

	var appointment model.Appointment
	err = mongo.AppointmentCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&appointment)
	if err != nil {
		c.JSON(404, gin.H{"error": "Appointment not found"})
		return
	}

	c.JSON(200, appointment)
}

func UpdateDetailAppointmentById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid appointment ID"})
		return
	}

	var appointment model.Appointment
	if err := c.BindJSON(&appointment); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"vehicle_brand":   appointment.VehicleBrand,
			"vehicle_model":   appointment.VehicleModel,
			"servicetype":     appointment.ServiceType,
			"appointmentdate": appointment.AppointmentDate,
			"pickupaddress":   appointment.PickupAddress,
			"dropoffaddress":  appointment.DropoffAddress,
			"status":          appointment.Status,
		},
	}

	_, err = mongo.AppointmentCollection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update appointment"})
		return
	}

	c.JSON(200, gin.H{"message": "Appointment updated"})
}

func DeleteDetailAppointmentById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid appointment ID"})
		return
	}

	_, err = mongo.AppointmentCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete appointment"})
		return
	}

	c.JSON(200, gin.H{"message": "Appointment deleted"})
}
