package service

import (
	"context"
	"strings"
	"time"
	mongo "vehicle/database"
	model "vehicle/model"
	smtpHelper "vehicle/smtp"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 獲取全部appointments
func GetAllAppointment(c *gin.Context) {
	userID := c.GetHeader("Authorization")

	if userID == "" {
		c.JSON(400, gin.H{"error": "User ID is required"})
		return
	}

	userID = strings.TrimPrefix(userID, "Bearer ")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid User ID format"})
		return
	}

	var appointments []model.Appointment
	filter := bson.M{"user_id": objID}
	cursor, err := mongo.AppointmentCollection.Find(context.Background(), filter)
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

// 新增appointment
func CreateNewAppointment(c *gin.Context) {
	userID := c.GetHeader("Authorization")

	if userID == "" {
		c.JSON(400, gin.H{"error": "User ID is required"})
		return
	}

	userID = strings.TrimPrefix(userID, "Bearer ")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid User ID format"})
		return
	}

	var appointment model.Appointment
	if err := c.BindJSON(&appointment); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	appointment.ID = primitive.NewObjectID()
	appointment.CreatedAt = time.Now()
	appointment.UserID = objID

	_, err = mongo.AppointmentCollection.InsertOne(context.Background(), appointment)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create appointment"})
		return
	}

	user, err := GetUserByID(userID)
	if err != nil {
		smtpHelper.SendReservationEmail(appointment, *user)
	}

	c.JSON(201, appointment)
}

// 以appointment ID 搜尋 detail
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

// 以appointment ID 更新 detail
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
			"comments":        appointment.Comments,
		},
	}

	_, err = mongo.AppointmentCollection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update appointment"})
		return
	}

	c.JSON(200, gin.H{"message": "Appointment updated"})
}

// 刪除指定appointId的資料
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

func GetClientAppointments(c *gin.Context) {
	//修正聚合
	var appointments []model.Appointment
	// 首先，找到所有角色为 "client" 的用户
	var users []model.User
	filter := bson.M{"role": "client"} // Filter appointments by client role
	cursor, err := mongo.UserCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &users); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode users"})
		return
	}
	var userIDs []primitive.ObjectID
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	//使用userIDs fileter appointment
	appointmentFilter := bson.M{"user_id": bson.M{"$in": userIDs}}
	appointmentCursor, err := mongo.AppointmentCollection.Find(context.Background(), appointmentFilter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch appointments"})
		return
	}
	defer appointmentCursor.Close(context.Background())

	if err := appointmentCursor.All(context.Background(), &appointments); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode appointments"})
		return
	}

	//修飾appintment顯示
	for i, appointment := range appointments {
		for _, user := range users {
			if appointment.UserID == user.ID {
				appointments[i].UserName = user.Username // 或其他合适的字段
			}
		}
	}

	c.JSON(200, appointments)
}

func GetBookedSlots(c *gin.Context) {
	// 解析请求中的日期参数（假设你通过查询参数传递起始日期和结束日期）
	startDate := c.Query("start_date") // 起始日期，例如 "2024-08-01"
	endDate := c.Query("end_date")     // 结束日期，例如 "2024-08-31"

	// 创建时间范围的过滤条件
	filter := bson.M{
		"appointment_date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	// 查找在指定日期范围内的所有预约
	cursor, err := mongo.AppointmentCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch appointments"})
		return
	}
	defer cursor.Close(context.Background())

	var appointments []model.Appointment
	if err := cursor.All(context.Background(), &appointments); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode appointments"})
		return
	}

	// 提取已预订的时段
	bookedSlots := make(map[string][]string)
	for _, appointment := range appointments {
		// 使用预约日期和时间来标记已预订的时段
		date := appointment.AppointmentDate
		bookedSlots[date] = append(bookedSlots[date], appointment.ServiceType) // 可以加入更多细节
	}

	c.JSON(200, bookedSlots)
}

func GetAvailableSlots(c *gin.Context) {
	date := c.Query("date")

	// 統計某日上下午
	morningFilter := bson.M{"appointment_date": date, "timeslot": "morning"}
	afternoonFilter := bson.M{"appointment_date": date, "timeslot": "afternoon"}

	morningCount, err := mongo.AppointmentCollection.CountDocuments(context.Background(), morningFilter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to count morning appointments"})
		return
	}

	afternoonCount, err := mongo.AppointmentCollection.CountDocuments(context.Background(), afternoonFilter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to count afternoon appointments"})
		return
	}

	// 查询 owner 设置的最大预约数
	var capacity model.TimeSlotCapacity
	capacityFilter := bson.M{"date": date}
	err = mongo.CapacityCollection.FindOne(context.Background(), capacityFilter).Decode(&capacity)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch time slot capacity"})
		return
	}

	// 計算剩餘
	availableSlots := map[string]int{
		"morning":   capacity.Morning - int(morningCount),
		"afternoon": capacity.Afternoon - int(afternoonCount),
	}

	c.JSON(200, availableSlots)
}
