package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stream struct {
	Id              primitive.ObjectID `json:"id" bson:"_id"`
	ServerIpAddress string             `json:"server_ip_address" bson:"server_ip_address"`
	Protocol        string             `json:"protocol" bson:"protocol"`
	Status          string             `json:"status" bson:"status"`
	ExternalId      string             `json:"external_id" bson:"external_id"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}
