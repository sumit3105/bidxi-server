package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Participant stores lobby participant state.
type Participant struct {
	UserID         primitive.ObjectID   `bson:"userId" json:"userId"`
	Username       string               `bson:"username" json:"username"`
	TeamName       string               `bson:"teamName" json:"teamName"`
	IplTeam        string               `bson:"iplTeam" json:"iplTeam"`
	PurseRemaining int                  `bson:"purseRemaining" json:"purseRemaining"`
	Squad          []primitive.ObjectID `bson:"squad,omitempty" json:"squad,omitempty"`
	JoinedAt       time.Time            `bson:"joinedAt" json:"joinedAt"`
}

// Lobby represents an auction lobby.
type Lobby struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	AdminID       primitive.ObjectID `bson:"adminId" json:"adminId"`
	InviteCode    string             `bson:"inviteCode" json:"inviteCode"`
	MaxPlayers    int                `bson:"maxPlayers" json:"maxPlayers"`
	Purse         int                `bson:"purse" json:"purse"`
	Status        string             `bson:"status" json:"status"` // waiting|active|completed
	Participants  []Participant      `bson:"participants,omitempty" json:"participants,omitempty"`
	AuctionRoomID primitive.ObjectID `bson:"auctionRoomId,omitempty" json:"auctionRoomId,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}
