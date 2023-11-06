package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	UserType string `json: "userType" bson:"usertype"`
}


//TODO: Change name to DeanSession
//TODO: Change WardenID to DeanID
type DeanSession struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	DeanId primitive.ObjectID `json:"deanId" bson:"dean_id"`
	Status string `json:"status" bson:"status"`
	BookerID primitive.ObjectID `json:"bookerId" bson:"booker_id"`
	SessionTime primitive.DateTime `json:"sessionTime" bson:"session_time"`
	ExpiresAt primitive.DateTime	`json:"expiresAt" bson:"expiresAt"`
}


