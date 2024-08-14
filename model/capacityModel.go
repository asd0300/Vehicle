package model

type TimeSlotCapacity struct {
	Date      string `bson:"date" json:"date"`
	Morning   int    `bson:"morning" json:"morning"`
	Afternoon int    `bson:"afternoon" json:"afternoon"`
}
