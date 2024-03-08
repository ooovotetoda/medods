package models

import "github.com/google/uuid"

type Authorization struct {
	UserGUID         uuid.UUID `bson:"_id"`
	RefreshTokenHash string    `bson:"refresh_token_hash"`
}
