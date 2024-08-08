package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type User struct {
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	r := gin.Default()
	clientOption := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(context.TODO(), clientOption)
	defer client.Disconnect(context.TODO())

	r.POST("/register", registerUser)
	r.Run(":4000")s
}

func registerUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hash := sha256.New()
	hash.Write([]byte(user.Password))
	user.Password = hex.EncodeToString(hash.Sum(nil))
	user.CreatedAt = time.Now()

	collection := client.Database("car-repair-system").Collection("users")
	_, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}
