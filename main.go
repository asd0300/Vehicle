package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
	database "vehicle/database"
	enviroment "vehicle/enviroment"
	router "vehicle/router"
	service "vehicle/service"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	enviroment.SetEnv()
	LogSetting()
}

func LogSetting() {
	currentDate := time.Now().Format("2006-01-02")

	logDir := "logs"
	logFileName := fmt.Sprintf("%s/%s.log", logDir, currentDate)

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err = os.Mkdir(logDir, 0755)
		if err != nil {
			log.Fatalf("Could not create log directory: %v", err)
		}
	}

	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to log to file, using default stderr: %v", err)
	}

	log.SetOutput(file)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.SetLevel(log.InfoLevel)
}

func main() {
	//初始化 mongo
	database.CreateMongoConnect()

	service.Setjwtkey(enviroment.JwtKey)

	// defer client.Disconnect(context.TODO())

	app := setupRouter()
	app.Run(":4000")
}

func setupRouter() *gin.Engine {
	app := gin.Default()
	// app.Use(timeoutMiddleware(5 * time.Second))
	app.Use(corsMiddleware())
	api := app.Group("api")
	{
		router.AddUserRoute(api)
		router.AppointmentRoute(api)
		router.AddCapacityRoute(api)
	}
	// app.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return app
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, userToken")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}
