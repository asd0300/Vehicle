package main

import (
	"net/http"
	"os"
	router "vehicle/router"
	service "vehicle/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var JwtKey []byte

func init() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	JwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
}

func main() {
	service.Setjwtkey(JwtKey)

	// defer client.Disconnect(context.TODO())

	app := setupRouter()
	app.Run(":4000")
}

func setupRouter() *gin.Engine {
	app := gin.Default()
	// app.Use(timeoutMiddleware(5 * time.Second))
	app.Use(corsMiddleware())
	api := app.Group("api")

	router.AddUserRoute(api)
	// app.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return app
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, userToken")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}
