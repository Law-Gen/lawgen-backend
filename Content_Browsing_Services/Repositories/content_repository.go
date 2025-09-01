package Repositories

import (
	"context"
	domain "lawgen/admin-service/Domain"

	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoContentRepository implements the Domain.ContentRepository interface.
// Its single responsibility is to interact with MongoDB for content metadata.
type mongoContentRepository struct {
	collection *mongo.Collection
}

func NewMongoContentRepository(db *mongo.Database) domain.IContentRepository {
	return &mongoContentRepository{collection: db.Collection("legal_contents")}
}

func (r *mongoContentRepository) Save(ctx context.Context, content *domain.Content) (string, error) {
	result, err := r.collection.InsertOne(ctx, content)
	if err != nil { return "", err }
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *mongoContentRepository) GetByID(ctx context.Context, id string) (*domain.Content, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil { return nil, errors.New("invalid ID format") }

	var content domain.Content
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&content)
	if err != nil {
		if err == mongo.ErrNoDocuments { return nil, domain.ErrNotFound }
		return nil, err
	}
	return &content, nil
}

func (r *mongoContentRepository) GetAll(ctx context.Context, page, limit int, search string) (*domain.PaginatedContentResponse, error) {
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	
	filter := bson.D{}
	if search != "" {
		filter = bson.D{{Key: "name", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}}
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil { return nil, err }
	defer cursor.Close(ctx)

	var contents []domain.Content
	if err = cursor.All(ctx, &contents); err != nil { return nil, err }

	totalItems, err := r.collection.CountDocuments(ctx, filter)
	if err != nil { return nil, err }

	return &domain.PaginatedContentResponse{
		Items:       contents,
		TotalItems:  int(totalItems),
		TotalPages:  (int(totalItems) + limit - 1) / limit,
		CurrentPage: page,
		PageSize:    limit,
	}, nil
}