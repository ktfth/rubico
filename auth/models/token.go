package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Token struct {
	ID        primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Token     string             `bson:"token"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId,omitempty"`
}
