package main

import (
	"net/http"
	database "vehicle/database"
	enviroment "vehicle/enviroment"
	router "vehicle/router"
	service "vehicle/service"

	"github.com/gin-gonic/gin"
)

func init() {
	enviroment.SetEnv()

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
