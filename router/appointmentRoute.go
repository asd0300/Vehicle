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

		book.DELETE("/appointments/:id", service.DeleteDetailAppointmentById)
		book.GET("/appointments/slot", service.GetBookedSlots)
	}
	protected := r.Group("/book")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/appointments/client", OwnerAppointmentsHandler, service.GetClientAppointments)
	protected.GET("/getAvailableSlots", service.GetAvailableSlots)
	protected.PUT("/appointments/:id", OwnerAppointmentsHandler, service.UpdateDetailAppointmentById)

	calendar := r.Group("/calendar")
	calendar.Use(middleware.AuthMiddleware())
	calendar.GET("", service.GetAppointmentResStatus)
	calendar.POST("limits", OwnerAppointmentsHandler, service.SettingSaveLimits)
	calendar.GET("limits", OwnerAppointmentsHandler, service.GetLimits)

}

func OwnerAppointmentsHandler(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	fmt.Print("message:This is owner appointments data")
}
