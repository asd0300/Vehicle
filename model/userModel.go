package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `json:"username" binding:"required"`
	Password  string             `json:"password" binding:"required"`
	Email     string             `json:"email" binding:"required,email"`
	Role      string             `json:"role"`
	Capacity  []TimeSlotCapacity `bson:"capacity,omitempty" json:"capacity"`
	CreatedAt time.Time          `json:"created_at"`
}
