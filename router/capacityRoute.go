package router

import (
	service "vehicle/service"

	"github.com/gin-gonic/gin"
)

func AddCapacityRoute(r *gin.RouterGroup) {
	capacity := r.Group("/capacity")
	capacity.POST("", service.SetMultipleTimeSlotCapacities)
	capacity.GET("", service.GetWeeklyCapacities)
}
