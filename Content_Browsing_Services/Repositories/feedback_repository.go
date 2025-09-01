package Repositories

import (
	"context"
	Domain "lawgen/admin-service/Domain"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoFeedbackRepository struct {
    collection *mongo.Collection
}

func NewMongoFeedbackRepository(db *mongo.Database) *MongoFeedbackRepository {
    return &MongoFeedbackRepository{
        collection: db.Collection("feedback"),
    }
}

func (r *MongoFeedbackRepository) Create(feedback *Domain.Feedback) error {
    if feedback.ID == "" {
        feedback.ID = uuid.NewString()
    }
    feedback.Timestamp = time.Now()
    _, err := r.collection.InsertOne(context.TODO(), feedback)
    return err
}

func (r *MongoFeedbackRepository) GetByID(id string) (*Domain.Feedback, error) {
    var f Domain.Feedback
    err := r.collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&f)
    if err != nil {
        return nil, err
    }
    return &f, nil
}

func (r *MongoFeedbackRepository) List() ([]Domain.Feedback, error) {
    cursor, err := r.collection.Find(context.TODO(), bson.M{})
    if err != nil {
        return nil, err
    }
    var feedbacks []Domain.Feedback
    if err := cursor.All(context.TODO(), &feedbacks); err != nil {
        return nil, err
    }
    return feedbacks, nil
}