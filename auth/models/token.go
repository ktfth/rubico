package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Token struct {
	ID        primitive.ObjectID `bson:"_id"`
	Token     string             `bson:"token"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	UserID    primitive.ObjectID `bson:"userId"`
}
