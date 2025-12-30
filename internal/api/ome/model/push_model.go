package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Push struct {
	Id              primitive.ObjectID `json:"id" bson:"_id"`
	StreamId        primitive.ObjectID `json:"stream_id" bson:"stream_id"`
	RtmpUrl         string             `json:"rtmp_url" bson:"rtmp_url"`
	Status          string             `json:"status" bson:"status"`
	ServerIpAddress string             `json:"server_ip_address" bson:"server_ip_address"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}
