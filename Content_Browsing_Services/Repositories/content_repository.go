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

// mongoContentRepository implements IContentRepository
type mongoContentRepository struct {
	collection *mongo.Collection
}

func NewMongoContentRepository(db *mongo.Database) domain.IContentRepository {
	return &mongoContentRepository{collection: db.Collection("legal_contents")}
}

// Save inserts a new content document
func (r *mongoContentRepository) Save(ctx context.Context, content *domain.Content) (string, error) {
	result, err := r.collection.InsertOne(ctx, content)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// GetByID fetches content by its unique ID
func (r *mongoContentRepository) GetByID(ctx context.Context, id string) (*domain.Content, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var content domain.Content
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&content)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &content, nil
}

// GetAllGroups returns paginated distinct groups
func (r *mongoContentRepository) GetAll(ctx context.Context, page, limit int) (*domain.PaginatedGroupResponse, error) {
	skip := (page - 1) * limit

	// Aggregate distinct group names with first document _id as group ID
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$group_name"},
			{Key: "first_id", Value: bson.D{{Key: "$first", Value: "$_id"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		{{Key: "$skip", Value: int64(skip)}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []domain.Group
	for cursor.Next(ctx) {
		var g struct {
			ID        string `bson:"first_id"`
			GroupName string `bson:"_id"`
		}
		if err := cursor.Decode(&g); err != nil {
			return nil, err
		}
		groups = append(groups, domain.Group{
			ID:        g.ID,
			GroupName: g.GroupName,
		})
	}

	// Count total distinct groups
	countPipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$group_name"}}}},
		{{Key: "$count", Value: "total"}},
	}
	countCursor, err := r.collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, err
	}
	defer countCursor.Close(ctx)

	totalItems := 0
	if countCursor.Next(ctx) {
		var c struct {
			Total int `bson:"total"`
		}
		_ = countCursor.Decode(&c)
		totalItems = c.Total
	}

	return &domain.PaginatedGroupResponse{
		Group:       groups,
		TotalItems:  totalItems,
		TotalPages:  (totalItems + limit - 1) / limit,
		CurrentPage: page,
		PageSize:    limit,
	}, nil
}

// GetContentsByGroup fetches paginated contents for a given group ID
func (r *mongoContentRepository) GetContentByGroup(ctx context.Context, groupID string, page, limit int) (*domain.PaginatedContentResponse, error) {
    objID, err := primitive.ObjectIDFromHex(groupID)
    if err != nil {
        return nil, errors.New("invalid group ID format")
    }

    filter := bson.M{"group_id": objID}

    opts := options.Find()
    opts.SetSkip(int64((page - 1) * limit))
    opts.SetLimit(int64(limit))

    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var contents []domain.Content
    if err := cursor.All(ctx, &contents); err != nil {
        return nil, err
    }

    totalItems, err := r.collection.CountDocuments(ctx, filter)
    if err != nil {
        return nil, err
    }

    return &domain.PaginatedContentResponse{
        Contents:    contents,
        TotalItems:  int(totalItems),
        TotalPages:  (int(totalItems) + limit - 1) / limit,
        CurrentPage: page,
        PageSize:    limit,
    }, nil
}


func (r *mongoContentRepository) FindOneByGroupName(ctx context.Context, groupName string) (*domain.Content, error) {
	var content domain.Content
	filter := bson.M{"group_name": groupName}

	err := r.collection.FindOne(ctx, filter).Decode(&content)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This is not a system error. It's a valid outcome meaning the group is new.
			// We return nil for the content and a specific domain error (or just nil).
			return nil, domain.ErrNotFound // Assuming you have a standard ErrNotFound
		}
		// This is a real database error
		return nil, err
	}

	return &content, nil
}


// Delete removes a content document and its associated file from Azure
func (r *mongoContentRepository) Delete(ctx context.Context, id string) error {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return errors.New("invalid ID format")
    }

    // Fetch content first (to know file URL for Azure delete)
    var content domain.Content
    err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&content)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return domain.ErrNotFound
        }
        return err
    }

    // Delete from MongoDB
    _, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        return err
    }
    return nil
}
