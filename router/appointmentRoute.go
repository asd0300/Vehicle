package router

import (
	"fmt"
	"net/http"
	"vehicle/middleware"
	service "vehicle/service"

	"github.com/gin-gonic/gin"
)

func AppointmentRoute(r *gin.RouterGroup) {
	book := r.Group("/book")
	{
		book.POST("/appointments", service.CreateNewAppointment)
		book.GET("/appointments", service.GetAllAppointment)
		book.GET("/appointments/:id", service.GetDetailAppointmentById)
		book.PUT("/appointments/:id", service.UpdateDetailAppointmentById)
		book.DELETE("/appointments/:id", service.DeleteDetailAppointmentById)
		book.GET("/appointments/slot", service.GetBookedSlots)
	}
	protected := r.Group("/book")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/appointments/client", OwnerAppointmentsHandler, service.GetClientAppointments)
}

func OwnerAppointmentsHandler(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	fmt.Print("message:This is owner appointments data")
}
