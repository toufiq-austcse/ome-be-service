package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"strings"
)

type OmeService struct {
	db *mongo.Database
}

func NewOmeService(db *mongo.Database) *OmeService {
	return &OmeService{
		db: db,
	}
}

func (s OmeService) GetStreamName(urlStr string) string {
	// URL: http://146.190.194.9:3333/app/test?direction=whip

	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	// Get path segments and return the last one
	path := parsedUrl.Path // e.g., "/app/test"
	// Remove leading slash and split by '/'
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) > 0 {
		return segments[len(segments)-1] // Returns "test"
	}

	return ""
}

func (s OmeService) Create(c context.Context) (interface{}, error) {
	collection := s.db.Collection("streams")
	document := map[string]interface{}{
		"status": "inactive",
	}
	result, err := collection.InsertOne(c, document)
	if err != nil {
		return nil, err
	}

	var createdStream map[string]interface{}
	err = collection.FindOne(c, map[string]interface{}{
		"_id": result.InsertedID,
	}).Decode(&createdStream)

	if err != nil {
		return nil, err
	}

	return createdStream, nil
}

func (s OmeService) UpdateStreamById(c context.Context, id primitive.ObjectID, updatedData map[string]interface{}) error {
	fmt.Println("Updating stream ID:", id)
	collection := s.db.Collection("streams")
	_, err := collection.UpdateByID(c, id, map[string]interface{}{
		"$set": updatedData,
	})
	return err
}

func (s OmeService) UpdateStatusByName(c context.Context, id string, status string) error {
	collection := s.db.Collection("streams")
	_, err := collection.UpdateByID(c, id, map[string]interface{}{
		"$set": map[string]interface{}{
			"status": status,
		},
	})
	return err
}
func (s *OmeService) FindByID(c context.Context, id primitive.ObjectID) (map[string]interface{}, error) {
	collection := s.db.Collection("streams")
	var stream map[string]interface{}
	err := collection.FindOne(c, map[string]interface{}{
		"_id": id,
	}).Decode(&stream)
	if err != nil {
		return nil, err
	}
	return stream, nil

}
