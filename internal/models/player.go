package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Player represents a real IPL player in the auction pool.
type Player struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Role        string             `bson:"role" json:"role"`               // BAT, BOWL, ALL, WK
	Nationality string             `bson:"nationality" json:"nationality"` // Indian, Overseas
	Category    string             `bson:"category" json:"category"`       // MARQUEE, BATSMAN, BOWLER, ALLROUNDER, WK
	BasePrice   int                `bson:"basePrice" json:"basePrice"`     // in Lakhs
	ImageUrl    string             `bson:"imageUrl" json:"imageUrl"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

// SoldRecord tracks which player was sold for how much.
type SoldRecord struct {
	PlayerID    primitive.ObjectID `bson:"playerId" json:"playerId"`
	PlayerName  string             `bson:"playerName" json:"playerName"`
	PlayerImage string             `bson:"playerImage" json:"playerImage"`
	PlayerRole  string             `bson:"playerRole" json:"playerRole"`
	WinnerID    primitive.ObjectID `bson:"winnerId" json:"winnerId"`
	TeamName    string             `bson:"teamName" json:"teamName"`
	IplTeam     string             `bson:"iplTeam" json:"iplTeam"`
	SoldPrice   int                `bson:"soldPrice" json:"soldPrice"`
	SoldAt      time.Time          `bson:"soldAt" json:"soldAt"`
}

// AuctionRoom stores the real-time state of an ongoing auction.
type AuctionRoom struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	LobbyID         primitive.ObjectID   `bson:"lobbyId" json:"lobbyId"`
	Status          string               `bson:"status" json:"status"` // idle, bidding, sold, completed
	CurrentPlayerID primitive.ObjectID   `bson:"currentPlayerId,omitempty" json:"currentPlayerId,omitempty"`
	CurrentBid      int                  `bson:"currentBid" json:"currentBid"`
	HighestBidderID primitive.ObjectID   `bson:"highestBidderId,omitempty" json:"highestBidderId,omitempty"`
	TimerSeconds    int                  `bson:"timerSeconds" json:"timerSeconds"`
	PlayerQueue     []primitive.ObjectID `bson:"playerQueue" json:"playerQueue"`
	UnsoldPlayers   []primitive.ObjectID `bson:"unsoldPlayers,omitempty" json:"unsoldPlayers,omitempty"`
	SoldPlayers     []SoldRecord         `bson:"soldPlayers" json:"soldPlayers"`
	IsHostPresent   bool                 `bson:"isHostPresent" json:"isHostPresent"`
}
