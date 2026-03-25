package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents the core user domain model stored in MongoDB.
type User struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username      string               `bson:"username" json:"username"`
	Email         string               `bson:"email" json:"email"`
	PasswordHash  string               `bson:"passwordHash" json:"-"` // never expose to client
	CreatedAt     time.Time            `bson:"createdAt" json:"createdAt"`
	LobbiesJoined []primitive.ObjectID `bson:"lobbiesJoined,omitempty" json:"lobbiesJoined,omitempty"`
}
