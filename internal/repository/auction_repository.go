package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"bidxi/internal/models"
	"bidxi/pkg/database"
)

const auctionCollection = "auctions"

type AuctionRepository interface {
	Create(ctx context.Context, auction *models.AuctionRoom) error
	GetByID(ctx context.Context, id string) (*models.AuctionRoom, error)
	GetByLobbyID(ctx context.Context, lobbyID string) (*models.AuctionRoom, error)
	Update(ctx context.Context, auction *models.AuctionRoom) error
}

type auctionRepository struct{}

func NewAuctionRepository() AuctionRepository {
	return &auctionRepository{}
}

func (r *auctionRepository) collection() *mongo.Collection {
	return database.Collection(auctionCollection)
}

func (r *auctionRepository) Create(ctx context.Context, auction *models.AuctionRoom) error {
	_, err := r.collection().InsertOne(ctx, auction)
	return err
}

func (r *auctionRepository) GetByID(ctx context.Context, id string) (*models.AuctionRoom, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid auction ID: %w", err)
	}

	var auction models.AuctionRoom
	err = r.collection().FindOne(ctx, bson.M{"_id": objectID}).Decode(&auction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &auction, nil
}

func (r *auctionRepository) GetByLobbyID(ctx context.Context, lobbyID string) (*models.AuctionRoom, error) {
	objectID, err := primitive.ObjectIDFromHex(lobbyID)
	if err != nil {
		return nil, fmt.Errorf("invalid lobby ID: %w", err)
	}

	var auction models.AuctionRoom
	err = r.collection().FindOne(ctx, bson.M{"lobbyId": objectID}).Decode(&auction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &auction, nil
}

func (r *auctionRepository) Update(ctx context.Context, auction *models.AuctionRoom) error {
	_, err := r.collection().ReplaceOne(ctx, bson.M{"_id": auction.ID}, auction)
	return err
}
