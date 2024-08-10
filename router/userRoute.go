package router

import (
	service "vehicle/service"

	"github.com/gin-gonic/gin"
)

func AddUserRoute(r *gin.RouterGroup) {

	reginster := r.Group("/user")
	reginster.POST("/login", service.LoginUser)
	reginster.POST("/register", service.RegisterUser)
	// reginster.Use(jwt.JWTAuthMiddleware())
	// reginster.POST("/login", service.)

}
