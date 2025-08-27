package Repositories

import "go.mongodb.org/mongo-driver/mongo"

type mongoEntityRepository struct {
	entityCollection *mongo.Collection
}

func NewMongoEntityRepository (db *mongo.Database) 