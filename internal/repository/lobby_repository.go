package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"bidxi/internal/models"
	"bidxi/pkg/database"
)

const lobbiesCollection = "lobbies"

// LobbyRepository defines data access for lobby objects.
type LobbyRepository interface {
	Create(ctx context.Context, lobby *models.Lobby) error
	GetByID(ctx context.Context, id string) (*models.Lobby, error)
	GetByUser(ctx context.Context, userID string) ([]*models.Lobby, error)
	GetByInviteCode(ctx context.Context, code string) (*models.Lobby, error)
	AddParticipant(ctx context.Context, lobbyID string, participant models.Participant) error
	RemoveParticipant(ctx context.Context, lobbyID string, userID string) error
	UpdateStatus(ctx context.Context, lobbyID string, status string) error
	UpdateParticipants(ctx context.Context, lobbyID string, participants []models.Participant) error
	EnsureIndexes(ctx context.Context) error
}

type lobbyRepository struct{}

// NewLobbyRepository creates a new LobbyRepository instance.
func NewLobbyRepository() LobbyRepository {
	return &lobbyRepository{}
}

func (r *lobbyRepository) collection() *mongo.Collection {
	return database.Collection(lobbiesCollection)
}

func (r *lobbyRepository) Create(ctx context.Context, lobby *models.Lobby) error {
	lobby.CreatedAt = time.Now()
	lobby.Status = "waiting"
	result, err := r.collection().InsertOne(ctx, lobby)
	if err != nil {
		return fmt.Errorf("failed to create lobby: %w", err)
	}
	lobby.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *lobbyRepository) GetByID(ctx context.Context, id string) (*models.Lobby, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid lobby ID: %w", err)
	}

	var lobby models.Lobby
	err = r.collection().FindOne(ctx, bson.M{"_id": objectID}).Decode(&lobby)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	return &lobby, nil
}

func (r *lobbyRepository) GetByUser(ctx context.Context, userID string) ([]*models.Lobby, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	filter := bson.M{
		"$or": []bson.M{
			{"adminId": objectID},
			{"participants.userId": objectID},
		},
	}

	opts := options.Find().SetSort(bson.M{"createdAt": -1}) // Newest first

	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find lobbies by user: %w", err)
	}
	defer cursor.Close(ctx)

	var lobbies []*models.Lobby
	if err = cursor.All(ctx, &lobbies); err != nil {
		return nil, fmt.Errorf("failed to decode lobbies: %w", err)
	}

	return lobbies, nil
}

func (r *lobbyRepository) GetByInviteCode(ctx context.Context, code string) (*models.Lobby, error) {
	var lobby models.Lobby
	err := r.collection().FindOne(ctx, bson.M{"inviteCode": code}).Decode(&lobby)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lobby by invite code: %w", err)
	}
	return &lobby, nil
}

func (r *lobbyRepository) AddParticipant(ctx context.Context, lobbyID string, participant models.Participant) error {
	objectID, err := primitive.ObjectIDFromHex(lobbyID)
	if err != nil {
		return fmt.Errorf("invalid lobby ID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$push": bson.M{"participants": participant}}
	res, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("lobby not found")
	}
	return nil
}

func (r *lobbyRepository) RemoveParticipant(ctx context.Context, lobbyID string, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(lobbyID)
	if err != nil {
		return fmt.Errorf("invalid lobby ID: %w", err)
	}

	participantID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid participant user ID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$pull": bson.M{"participants": bson.M{"userId": participantID}}}
	_, err = r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}
	return nil
}

func (r *lobbyRepository) UpdateStatus(ctx context.Context, lobbyID string, status string) error {
	objectID, err := primitive.ObjectIDFromHex(lobbyID)
	if err != nil {
		return fmt.Errorf("invalid lobby ID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"status": status}}
	res, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update lobby status: %w", err)
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("lobby not found")
	}
	return nil
}

func (r *lobbyRepository) UpdateParticipants(ctx context.Context, lobbyID string, participants []models.Participant) error {
	objectID, err := primitive.ObjectIDFromHex(lobbyID)
	if err != nil {
		return fmt.Errorf("invalid lobby ID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"participants": participants}}
	_, err = r.collection().UpdateOne(ctx, filter, update)
	return err
}

func (r *lobbyRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "inviteCode", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := r.collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to ensure lobby indexes: %w", err)
	}
	return nil
}
