package service

import (
	"context"
	"net/http"
	"strings"
	"time"
	mongovehicle "vehicle/database"
	"vehicle/kafkahelper"
	model "vehicle/model"
	smtpHelper "vehicle/smtp"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	cursor, err := mongovehicle.AppointmentCollection.Find(context.Background(), filter)
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

	// appointment.ID = primitive.NewObjectID()
	// appointment.CreatedAt = time.Now()
	// appointment.UserID = objID

	// _, err = mongovehicle.AppointmentCollection.InsertOne(context.Background(), appointment)
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": "Failed to create appointment"})
	// 	return
	// }

	// Start a session for transaction
	session, err := mongovehicle.Client.StartSession()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start session"})
		return
	}
	defer session.EndSession(context.Background())

	// Create a transaction
	err = session.StartTransaction()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Transaction block
	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		// 1. 查找并更新相应日期的 Reserved 数量
		filter := bson.M{"day": appointment.AppointmentDate, "resdue_appointment": bson.M{"$gt": 0}}
		update := bson.M{
			"$inc": bson.M{"reserved": 1, "resdue_appointment": -1},
		}
		options := options.FindOneAndUpdate().SetReturnDocument(options.After)

		var updatedCalendar model.CalendarData
		err = mongovehicle.CalendarCollection.FindOneAndUpdate(sc, filter, update, options).Decode(&updatedCalendar)
		if err != nil {
			session.AbortTransaction(sc)
			if err == mongo.ErrNoDocuments {
				c.JSON(400, gin.H{"error": "No available appointment slots on the selected date"})
			} else {
				c.JSON(500, gin.H{"error": "Failed to update calendar"})
			}
			return err
		}

		// 2. 创建新的预约
		appointment.ID = primitive.NewObjectID()
		appointment.CreatedAt = time.Now()
		appointment.UserID = objID

		_, err = mongovehicle.AppointmentCollection.InsertOne(sc, appointment)
		if err != nil {
			session.AbortTransaction(sc)
			c.JSON(500, gin.H{"error": "Failed to create appointment"})
			return err
		}

		// 3. 提交事务
		if err = session.CommitTransaction(sc); err != nil {
			c.JSON(500, gin.H{"error": "Failed to commit transaction"})
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	//寄信通知獲得使用者資訊
	user, err := GetUserByID(userID)
	if err != nil {
		log.Errorf("Error getting user by ID: %v, fail to Send Mail", err)
	}

	//是否使用kfaka
	if kafkahelper.IsKafkaOn() {
		err := kafkahelper.SendReservationEmailKafkaProducer(appointment, *user)
		if err != nil {
			log.Warnf("Failed to send email via Kafka: %v, fallback to direct email sending", err)
			smtpHelper.SendReservationEmail(appointment, *user)
		}
	} else {
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
	err = mongovehicle.AppointmentCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&appointment)
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

	_, err = mongovehicle.AppointmentCollection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
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

	var appointment model.Appointment
	err = mongovehicle.AppointmentCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&appointment)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to find appointment"})
		return
	}

	_, err = mongovehicle.AppointmentCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete appointment"})
		return
	}

	// 更新 CalendarData，原子增加 ResdueAppointment 和减少 Reserved
	filter := bson.M{"day": appointment.AppointmentDate}
	update := bson.M{
		"$inc": bson.M{
			"reserved":           -1,
			"resdue_appointment": 1,
		},
	}

	result := mongovehicle.CalendarCollection.FindOneAndUpdate(context.Background(), filter, update)
	if result.Err() != nil {
		c.JSON(500, gin.H{"error": "Failed to update calendar data"})
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
	cursor, err := mongovehicle.UserCollection.Find(context.Background(), filter)
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
	appointmentCursor, err := mongovehicle.AppointmentCollection.Find(context.Background(), appointmentFilter)
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
	cursor, err := mongovehicle.AppointmentCollection.Find(context.Background(), filter)
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

	morningCount, err := mongovehicle.AppointmentCollection.CountDocuments(context.Background(), morningFilter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to count morning appointments"})
		return
	}

	afternoonCount, err := mongovehicle.AppointmentCollection.CountDocuments(context.Background(), afternoonFilter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to count afternoon appointments"})
		return
	}

	// 查询 owner 设置的最大预约数
	var capacity model.TimeSlotCapacity
	capacityFilter := bson.M{"date": date}
	err = mongovehicle.CapacityCollection.FindOne(context.Background(), capacityFilter).Decode(&capacity)
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

// 獲取可使用預約
func GetAppointmentResStatus(c *gin.Context) {
	var calendarData []model.CalendarData
	cursor, err := mongovehicle.CalendarCollection.Find(context.Background(), bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var data model.CalendarData
		err := cursor.Decode(&data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode data"})
			return
		}
		calendarData = append(calendarData, data)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	c.JSON(http.StatusOK, calendarData)
}

// owner 儲存 calendar limit
func SettingSaveLimits(c *gin.Context) {
	var rawLimits []model.CalendarData
	if err := c.BindJSON(&rawLimits); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	for _, limit := range rawLimits {
		filter := bson.M{"day": limit.Day}
		update := bson.M{
			"$set": bson.M{
				"limit_appointment":  limit.LimitAppointment,
				"resdue_appointment": limit.ResdueAppointment,
			},
		}
		_, err := mongovehicle.CalendarCollection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save limits"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Limits saved successfully"})
}

func GetLimits(c *gin.Context) {
	var limits []model.CalendarData

	cursor, err := mongovehicle.CalendarCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch limits"})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &limits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse limits"})
		return
	}

	c.JSON(http.StatusOK, limits)
}
