package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	router "vehicle/router"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	app := gin.Default()
	// app.Use(timeoutMiddleware(5 * time.Second))
	app.Use(corsMiddleware())
	api := app.Group("api")
	{
		router.AddUserRoute(api)
		router.AppointmentRoute(api)
		router.AddCapacityRoute(api)
	}
	return app
}

func TestCreateNewAppointmentRaceCondition(t *testing.T) {
	r := setupRouter()

	const concurrentRequests = 1
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	// appointmentDate := "2024-08-20"

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()

			appointmentData := `{"vehicle_brand":"Toyota","vehicle_model":"Camry","service_type":"maintenance","appointment_date":"2024-08-20","morning_slot":"","afternoon_slot":"","pickup_address":"2","dropoff_address":"1"}`

			req, _ := http.NewRequest("POST", "/api/book/appointments", strings.NewReader(appointmentData))
			req.Header.Set("Authorization", "Bearer 66b9a601437e6042832ed115")
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("Expected status 201, but got %d", w.Code)
			}
		}()
	}

	wg.Wait()
}
