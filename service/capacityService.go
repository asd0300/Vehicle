package service

import (
	"context"
	"net/http"
	"time"
	mongovehicle "vehicle/database"
	model "vehicle/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetMultipleTimeSlotCapacities(c *gin.Context) {
	var capacities []model.TimeSlotCapacity
	if err := c.ShouldBindJSON(&capacities); err != nil {
		c.JSON(400, gin.H{"error": "Invalid capacity data"})
		return
	}

	for _, capacity := range capacities {
		filter := bson.M{"date": capacity.Date}
		update := bson.M{
			"$set": bson.M{
				"morning":   capacity.Morning,
				"afternoon": capacity.Afternoon,
			},
		}

		_, err := mongovehicle.CapacityCollection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to set time slot capacity"})
			return
		}
	}

	c.JSON(200, gin.H{"message": "Time slot capacities set successfully"})
}

func GetWeeklyCapacities(c *gin.Context) {
	startDate := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	endDate := startDate.AddDate(0, 0, 7)

	filter := bson.M{
		"date": bson.M{
			"$gte": startDate.Format("2006-01-02"),
			"$lt":  endDate.Format("2006-01-02"),
		},
	}

	cursor, err := mongovehicle.CapacityCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch capacities"})
		return
	}
	defer cursor.Close(context.Background())

	var capacities []model.TimeSlotCapacity
	if err := cursor.All(context.Background(), &capacities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode capacities"})
		return
	}

	c.JSON(http.StatusOK, capacities)
}

// func GetAvailableSlots(c *gin.Context) {
// 	date := c.Query("date")
// 	if date == "" {
// 		c.JSON(400, gin.H{"error": "Date is required"})
// 		return
// 	}

// 	var capacity model.TimeSlotCapacity
// 	filter := bson.M{"date": date}
// 	err := mongovehicle.CapacityCollection.FindOne(context.Background(), filter).Decode(&capacity)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			c.JSON(404, gin.H{"error": "No capacity settings found for the given date"})
// 		} else {
// 			c.JSON(500, gin.H{"error": "Failed to fetch capacity"})
// 		}
// 		return
// 	}

// 	timeSlots := map[string][]string{
// 		"morning":   {"08:00", "09:00", "10:00", "11:00"},
// 		"afternoon": {"13:00", "14:00", "15:00", "16:00"},
// 	}

// 	availableSlots := map[string][]string{
// 		"morning":   timeSlots["morning"][:capacity.Morning],
// 		"afternoon": timeSlots["afternoon"][:capacity.Afternoon],
// 	}

// 	c.JSON(200, availableSlots)
// }
