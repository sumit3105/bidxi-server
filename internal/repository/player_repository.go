package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"bidxi/internal/models"
	"bidxi/pkg/database"
)

const playersCollection = "players"

type PlayerRepository interface {
	Create(ctx context.Context, player *models.Player) error
	GetAll(ctx context.Context) ([]*models.Player, error)
	GetByID(ctx context.Context, id string) (*models.Player, error)
	GetMany(ctx context.Context, ids []string) ([]*models.Player, error)
	Count(ctx context.Context) (int64, error)
	ClearAll(ctx context.Context) error
}

type playerRepository struct{}

func NewPlayerRepository() PlayerRepository {
	return &playerRepository{}
}

func (r *playerRepository) collection() *mongo.Collection {
	return database.Collection(playersCollection)
}

func (r *playerRepository) Create(ctx context.Context, player *models.Player) error {
	player.CreatedAt = time.Now()
	_, err := r.collection().InsertOne(ctx, player)
	return err
}

func (r *playerRepository) GetAll(ctx context.Context) ([]*models.Player, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []*models.Player
	if err = cursor.All(ctx, &players); err != nil {
		return nil, err
	}
	return players, nil
}

func (r *playerRepository) GetByID(ctx context.Context, id string) (*models.Player, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid player ID: %w", err)
	}

	var player models.Player
	err = r.collection().FindOne(ctx, bson.M{"_id": objectID}).Decode(&player)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *playerRepository) GetMany(ctx context.Context, ids []string) ([]*models.Player, error) {
	if len(ids) == 0 {
		return []*models.Player{}, nil
	}

	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			objectIDs = append(objectIDs, oid)
		}
	}

	cursor, err := r.collection().Find(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []*models.Player
	if err = cursor.All(ctx, &players); err != nil {
		return nil, err
	}
	return players, nil
}

func (r *playerRepository) Count(ctx context.Context) (int64, error) {
	return r.collection().CountDocuments(ctx, bson.M{})
}

func (r *playerRepository) ClearAll(ctx context.Context) error {
	_, err := r.collection().DeleteMany(ctx, bson.M{})
	return err
}
