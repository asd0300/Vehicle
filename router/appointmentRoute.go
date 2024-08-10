package router

import (
	service "vehicle/service"

	"github.com/gin-gonic/gin"
)

func AppointmentRoute(r *gin.RouterGroup) {

	r.POST("/appointments", service.CreateNewAppointment)
	r.GET("/appointments", service.GetAllAppointment)
	r.GET("/appointments/:id", service.GetDetailAppointmentById)
	r.PUT("/appointments/:id", service.UpdateDetailAppointmentById)
	r.DELETE("/appointments/:id", service.DeleteDetailAppointmentById)
}
