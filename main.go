package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

var jwtKey []byte

func init() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
}

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

	r.POST("/login", loginUser)
	r.POST("/register", registerUser)
	r.Run(":4000")
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func loginUser(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	collection := client.Database("car-repair-system").Collection("users")
	err := collection.FindOne(context.TODO(), bson.M{"username": loginRequest.Username}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Compare passwords
	hash := sha256.New()
	hash.Write([]byte(loginRequest.Password))
	if user.Password != hex.EncodeToString(hash.Sum(nil)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(jwtKey)

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
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
