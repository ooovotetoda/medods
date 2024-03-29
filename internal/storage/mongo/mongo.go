package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"medods/internal/storage/models"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client *mongo.Client
}

func New(uri string, timeout time.Duration) (*Storage, error) {
	const op = "storage.mongo.New"

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{client: client}, nil
}

func (s *Storage) Close(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.client.Disconnect(ctx)
}

func (s *Storage) SaveRefreshTokenHash(authToken *models.Authorization, timeout time.Duration) error {
	const op = "storage.mongoclient.SaveRefreshToken"

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	collection := s.client.Database("medodsDatabase").Collection("authorization")
	_, err := collection.InsertOne(ctx, authToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) VerifyRefreshTokenHash(token string, timeout time.Duration) (*models.Authorization, error) {
	const op = "storage.mongoclient.VerifyRefreshTokenHash"

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	collection := s.client.Database("medodsDatabase").Collection("authorization")

	var authToken models.Authorization
	err := collection.FindOne(ctx, bson.M{"refresh_token_hash": bson.M{"$exists": true}}).Decode(&authToken)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find token: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(authToken.RefreshTokenHash), []byte(token))
	if err != nil {
		return nil, fmt.Errorf("%s: invalid token: %w", op, err)
	}

	return &authToken, nil
}
