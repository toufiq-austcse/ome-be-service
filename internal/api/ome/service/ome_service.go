package service

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/ome/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OmeService struct {
	streamCollection *mongo.Collection
	pushCollection   *mongo.Collection
}

func NewOmeService(db *mongo.Database) *OmeService {
	streamCollection := db.Collection("streams")
	pushCollection := db.Collection("pushes")

	return &OmeService{
		streamCollection: streamCollection,
		pushCollection:   pushCollection,
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

func (s OmeService) CreateStream(
	ctx context.Context, model *model.Stream) (*model.Stream, error) {
	model.Id = primitive.NewObjectID()
	currentTime := time.Now()
	model.CreatedAt = currentTime
	model.UpdatedAt = currentTime

	_, err := s.streamCollection.InsertOne(ctx, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (s OmeService) GetStreamByExternalId(ctx context.Context, externalId string) (*model.Stream, error) {
	var result model.Stream

	filter := bson.M{"external_id": externalId}
	err := s.streamCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (s OmeService) FindStreamById(c context.Context, id primitive.ObjectID) (*model.Stream, error) {
	var result model.Stream

	filter := bson.M{"_id": id}
	err := s.streamCollection.FindOne(c, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}
func (s OmeService) UpdateStreamByID(ctx context.Context, id primitive.ObjectID, update bson.M) (*model.Stream, error) {
	update["updated_at"] = time.Now()

	result, err := s.streamCollection.UpdateByID(
		ctx,
		id,
		bson.M{"$set": update},
	)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, nil
	}

	updatedStream, err := s.FindStreamById(ctx, id)
	if err != nil {
		return nil, err
	}
	return updatedStream, nil
}

func (s OmeService) CreatePush(
	ctx context.Context, model *model.Push) (*model.Push, error) {
	model.Id = primitive.NewObjectID()
	currentTime := time.Now()
	model.CreatedAt = currentTime
	model.UpdatedAt = currentTime

	_, err := s.pushCollection.InsertOne(ctx, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (s OmeService) FindPushByStreamIdAndStatus(ctx context.Context, id primitive.ObjectID, status string) (*model.Push, error) {
	var result model.Push

	filter := bson.M{"stream_id": id, "status": status}
	err := s.pushCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}
func (s OmeService) UpdatePushByID(ctx context.Context, id primitive.ObjectID, update bson.M) (*model.Push, error) {
	update["updated_at"] = time.Now()

	result, err := s.pushCollection.UpdateByID(
		ctx,
		id,
		bson.M{"$set": update},
	)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, nil
	}
	updatedPush, err := s.FindPushById(ctx, id)
	if err != nil {
		return nil, err
	}
	return updatedPush, nil
}

func (s OmeService) FindPushById(ctx context.Context, id primitive.ObjectID) (*model.Push, error) {
	var result model.Push

	filter := bson.M{"_id": id}
	err := s.pushCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil

}
