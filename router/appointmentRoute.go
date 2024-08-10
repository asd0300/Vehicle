package router

import (
	service "vehicle/service"

	"github.com/gin-gonic/gin"
)

func AppointmentRoute(r *gin.RouterGroup) {
	book := r.Group("/book")
	book.POST("/appointments", service.CreateNewAppointment)
	book.GET("/appointments", service.GetAllAppointment)
	book.GET("/appointments/:id", service.GetDetailAppointmentById)
	book.PUT("/appointments/:id", service.UpdateDetailAppointmentById)
	book.DELETE("/appointments/:id", service.DeleteDetailAppointmentById)
}
