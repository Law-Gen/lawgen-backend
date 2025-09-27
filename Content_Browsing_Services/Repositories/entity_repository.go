package Repositories

import (
	"context"
	"errors"
	
	domain "lawgen/admin-service/Domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoLegalEntityRepository struct {
	collection *mongo.Collection
}

func NewMongoEntityRepository (db *mongo.Database) domain.ILegalEntityRepository {
	return &mongoLegalEntityRepository {
		collection: db.Collection("legal_entity"),
	}
}

func (r *mongoLegalEntityRepository) Save(ctx context.Context, entity *domain.LegalEntity) (*domain.LegalEntity, error) {
	entity.ID = ""
	result, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return nil, err
	}

	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		entity.ID = oid.Hex()
	}

	return entity, nil
}

func (r *mongoLegalEntityRepository) GetByID(ctx context.Context, id string) (*domain.LegalEntity, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var entity domain.LegalEntity

	if err := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("entity not found")
		}

		return nil, err
	}

	return &entity, nil
}

func (r *mongoLegalEntityRepository) GetAll(ctx context.Context, page, limit int, search string) (*domain.PaginatedLegalEntityResponse, error) {
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

	var entities []domain.LegalEntity
	if err = cursor.All(ctx, &entities); err != nil {
		return nil, err
	}


	totalItems, err := r.collection.CountDocuments(ctx, filter)
	if err != nil { return  nil, err }

	return &domain.PaginatedLegalEntityResponse{
		Items: entities,
		TotalItems: int(totalItems),
		TotalPages: (int(totalItems) + limit - 1) / limit,
		CurrentPage: page,
		PageSize: limit,
	}, nil
}

func (r *mongoLegalEntityRepository) Update(ctx context.Context, id string, entity *domain.LegalEntity) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": objID}, entity)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("entity not found")
	}
	return nil
}

func (r *mongoLegalEntityRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return  err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id":objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("entity not found")
	}

	return nil
}