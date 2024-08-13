package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var AppointmentCollection *mongo.Collection

func init() {
	CreateMongoConnect()
}

func CreateMongoConnect() {
	clientOption := options.Client().ApplyURI("mongodb://localhost:27017")
	Client, err := mongo.Connect(context.TODO(), clientOption)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// 設定 ping 的超時時間為 5 秒
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	} else {
		fmt.Println("Successfully connected to MongoDB!")
	}
	AppointmentCollection = Client.Database("car-repair-system").Collection("appointments")
}
