package database

import (
	"context"

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
	Client, _ = mongo.Connect(context.TODO(), clientOption)
	AppointmentCollection = Client.Database("car-repair-system").Collection("appointments")
}
