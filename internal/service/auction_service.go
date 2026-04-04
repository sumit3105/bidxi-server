package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"bidxi/internal/models"
	"bidxi/internal/repository"
	"bidxi/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuctionService interface {
	InitializeAuction(ctx context.Context, lobbyID string) (*models.AuctionRoom, error)
	GetAuctionRoom(ctx context.Context, lobbyID string) (interface{}, error)
	PlaceBid(ctx context.Context, lobbyID string, userID string, amount int) error
	ToggleAuctionState(ctx context.Context, lobbyID string, action string) error
	HandleHostPresenceChange(ctx context.Context, lobbyID string, isPresent bool) error
}

type auctionService struct {
	auctionRepo repository.AuctionRepository
	lobbyRepo   repository.LobbyRepository
	playerRepo  repository.PlayerRepository
	timers      map[string]*time.Timer
	timerMutex  sync.Mutex
}

func NewAuctionService(
	auctionRepo repository.AuctionRepository,
	lobbyRepo repository.LobbyRepository,
	playerRepo repository.PlayerRepository,
) AuctionService {
	return &auctionService{
		auctionRepo: auctionRepo,
		lobbyRepo:   lobbyRepo,
		playerRepo:  playerRepo,
		timers:      make(map[string]*time.Timer),
	}
}

func (s *auctionService) InitializeAuction(ctx context.Context, lobbyID string) (*models.AuctionRoom, error) {
	allPlayers, err := s.playerRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// 1. Group by category (normalized uppercase)
	categories := []string{"MARQUEE", "WK", "BATSMAN", "ALLROUNDER", "BOWLER"}
	grouped := make(map[string][]primitive.ObjectID)
	usedIDs := make(map[primitive.ObjectID]bool)

	for _, p := range allPlayers {
		cat := strings.ToUpper(p.Category)
		grouped[cat] = append(grouped[cat], p.ID)
	}

	// 2. Shuffle each group and concatenate
	finalQueue := []primitive.ObjectID{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, cat := range categories {
		ids := grouped[cat]
		if len(ids) == 0 {
			continue
		}
		r.Shuffle(len(ids), func(i, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})
		for _, id := range ids {
			finalQueue = append(finalQueue, id)
			usedIDs[id] = true
		}
	}

	// 3. Catch-all for any players not in the above categories
	extraIDs := []primitive.ObjectID{}
	for _, p := range allPlayers {
		if !usedIDs[p.ID] {
			extraIDs = append(extraIDs, p.ID)
		}
	}
	if len(extraIDs) > 0 {
		r.Shuffle(len(extraIDs), func(i, j int) {
			extraIDs[i], extraIDs[j] = extraIDs[j], extraIDs[i]
		})
		finalQueue = append(finalQueue, extraIDs...)
	}

	log.Printf("✨ Initialized auction with %d players in queue", len(finalQueue))

	lid, _ := primitive.ObjectIDFromHex(lobbyID)
	room := &models.AuctionRoom{
		ID:          primitive.NewObjectID(),
		LobbyID:     lid,
		Status:      "idle", // Auction starts in idle state
		PlayerQueue: finalQueue,
		SoldPlayers: []models.SoldRecord{},
	}

	if err := s.auctionRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *auctionService) ToggleAuctionState(ctx context.Context, lobbyID string, action string) error {
	room, err := s.auctionRepo.GetByLobbyID(ctx, lobbyID)
	if err != nil {
		return err
	}

	if room == nil {
		if action == "start" {
			// Initialize room if it doesn't exist yet
			room, err = s.InitializeAuction(ctx, lobbyID)
			if err != nil {
				return fmt.Errorf("failed to initialize auction: %w", err)
			}
		} else {
			return fmt.Errorf("auction not started yet")
		}
	}

	switch action {
	case "start":
		if room.Status != "idle" {
			return fmt.Errorf("auction already started")
		}
		return s.SetNextPlayer(ctx, lobbyID)

	case "pause":
		if room.Status != "bidding" {
			return fmt.Errorf("cannot pause in state: %s", room.Status)
		}
		s.stopTimer(lobbyID)
		room.Status = "paused"

	case "resume":
		if room.Status != "paused" {
			return fmt.Errorf("auction not paused")
		}
		room.Status = "bidding"
		s.startTimer(lobbyID)

	case "next":
		if room.Status != "sold" && room.Status != "unsold" && room.Status != "idle" {
			return fmt.Errorf("cannot go to next player in current state: %s", room.Status)
		}
		return s.SetNextPlayer(ctx, lobbyID)

	case "retry_unsold":
		if len(room.UnsoldPlayers) == 0 {
			return fmt.Errorf("no unsold players to retry")
		}
		room.PlayerQueue = append(room.PlayerQueue, room.UnsoldPlayers...)
		room.UnsoldPlayers = nil
		room.Status = "idle" // Reset from "completed" so SetNextPlayer can proceed
		if err := s.auctionRepo.Update(ctx, room); err != nil {
			return fmt.Errorf("failed to save retry state: %w", err)
		}
		return s.SetNextPlayer(ctx, lobbyID)

	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	if err := s.auctionRepo.Update(ctx, room); err != nil {
		return err
	}

	// Fetch upcoming players to ensure the list is populated on resume/load
	upcomingPlayerIDs := []string{}
	for _, id := range room.PlayerQueue {
		upcomingPlayerIDs = append(upcomingPlayerIDs, id.Hex())
	}
	upcomingPlayers, _ := s.playerRepo.GetMany(ctx, upcomingPlayerIDs)

	player, _ := s.playerRepo.GetByID(ctx, room.CurrentPlayerID.Hex())

	websocket.GlobalHub.BroadcastToLobby(lobbyID, "AUCTION_STATE_UPDATED", map[string]interface{}{
		"room":     room,
		"player":   player,
		"upcoming": upcomingPlayers,
	})
	return nil
}

func (s *auctionService) SetNextPlayer(ctx context.Context, lobbyID string) error {
	room, err := s.auctionRepo.GetByLobbyID(ctx, lobbyID)
	if err != nil || room == nil {
		return fmt.Errorf("room not found")
	}

	if len(room.PlayerQueue) == 0 {
		room.Status = "completed"
		s.auctionRepo.Update(ctx, room)
		websocket.GlobalHub.BroadcastToLobby(lobbyID, "AUCTION_COMPLETED", room)
		return nil
	}

	nextPlayerID := room.PlayerQueue[0]
	room.PlayerQueue = room.PlayerQueue[1:]
	room.CurrentPlayerID = nextPlayerID

	// Important: Save the queue update to DB immediately to avoid infinite loop on skip
	if err := s.auctionRepo.Update(ctx, room); err != nil {
		return err
	}

	player, err := s.playerRepo.GetByID(ctx, nextPlayerID.Hex())
	if err != nil || player == nil {
		fmt.Printf("⚠️ Warning: Player %s not found. Skipping.\n", nextPlayerID.Hex())
		return s.SetNextPlayer(ctx, lobbyID)
	}

	upcomingPlayerIDs := []string{}
	for _, id := range room.PlayerQueue {
		upcomingPlayerIDs = append(upcomingPlayerIDs, id.Hex())
	}
	upcomingPlayers, _ := s.playerRepo.GetMany(ctx, upcomingPlayerIDs)

	room.CurrentBid = player.BasePrice
	room.HighestBidderID = primitive.NilObjectID
	room.TimerSeconds = 10
	room.Status = "bidding"

	if err := s.auctionRepo.Update(ctx, room); err != nil {
		return err
	}

	websocket.GlobalHub.BroadcastToLobby(lobbyID, "PLAYER_UP", map[string]interface{}{
		"player":   player,
		"room":     room,
		"upcoming": upcomingPlayers,
	})

	s.startTimer(lobbyID)
	return nil
}

func (s *auctionService) PlaceBid(ctx context.Context, lobbyID string, userID string, amount int) error {
	room, err := s.auctionRepo.GetByLobbyID(ctx, lobbyID)
	if err != nil || room == nil || room.Status != "bidding" {
		return fmt.Errorf("bidding not open")
	}

	minIncrement := 10
	if room.CurrentBid > 100 && room.CurrentBid <= 1000 {
		minIncrement = 20
	} else if room.CurrentBid > 1000 {
		minIncrement = 50
	}
	if amount < room.CurrentBid+minIncrement && !(room.HighestBidderID.IsZero() && amount == room.CurrentBid) {
		return fmt.Errorf("minimum increment is %dL", minIncrement)
	}

	uid, _ := primitive.ObjectIDFromHex(userID)
	if !room.HighestBidderID.IsZero() && room.HighestBidderID == uid {
		return fmt.Errorf("you are already the highest bidder")
	}

	lobby, _ := s.lobbyRepo.GetByID(ctx, lobbyID)
	for _, p := range lobby.Participants {
		if p.UserID == uid {
			if len(p.Squad) >= 25 {
				return fmt.Errorf("squad is full (max 25)")
			}

			// Validate constraints
			squadPlayerIDs := make([]string, 0, len(p.Squad))
			for _, sqID := range p.Squad {
				squadPlayerIDs = append(squadPlayerIDs, sqID.Hex())
			}
			squadPlayers, _ := s.playerRepo.GetMany(ctx, squadPlayerIDs)

			overseasCount, wkCount, bowlCount, batCount := 0, 0, 0, 0
			for _, sqP := range squadPlayers {
				if sqP.Nationality != "Indian" {
					overseasCount++
				}
				role := strings.ToUpper(sqP.Role)
				if strings.Contains(role, "WK") {
					wkCount++
				} else if strings.Contains(role, "BOWL") {
					bowlCount++
				} else if strings.Contains(role, "BAT") {
					batCount++
				}
			}

			currentPlayer, err := s.playerRepo.GetByID(ctx, room.CurrentPlayerID.Hex())
			if err == nil && currentPlayer != nil {
				if currentPlayer.Nationality != "Indian" && overseasCount >= 8 {
					return fmt.Errorf("maximum 8 overseas players allowed")
				}

				// Projected counts if this player is bought
				if currentPlayer.Role == "WK" {
					wkCount++
				} else if currentPlayer.Role == "BOWL" || currentPlayer.Role == "BOWLER" {
					bowlCount++
				} else if currentPlayer.Role == "BAT" || currentPlayer.Role == "BATSMAN" {
					batCount++
				}
			}

			missingWK := 0
			if 1-wkCount > 0 {
				missingWK = 1 - wkCount
			}
			missingBowl := 0
			if 5-bowlCount > 0 {
				missingBowl = 5 - bowlCount
			}
			missingBat := 0
			if 5-batCount > 0 {
				missingBat = 5 - batCount
			}
			totalMissingRoles := missingWK + missingBowl + missingBat

			slotsTo15 := 15 - (len(p.Squad) + 1)
			if slotsTo15 < 0 {
				slotsTo15 = 0
			}

			requiredFutureBuys := slotsTo15
			if totalMissingRoles > requiredFutureBuys {
				requiredFutureBuys = totalMissingRoles
			}

			reservedPurse := requiredFutureBuys * 200
			remainingPurseAfterBid := (p.PurseRemaining * 100) - amount

			if remainingPurseAfterBid < reservedPurse {
				return fmt.Errorf("insufficient purse: must reserve %d Cr for %d required slots", reservedPurse/100, requiredFutureBuys)
			}

			if amount > p.PurseRemaining*100 {
				return fmt.Errorf("insufficient purse")
			}
			break
		}
	}

	room.HighestBidderID = uid
	room.CurrentBid = amount
	room.TimerSeconds = 10
	s.auctionRepo.Update(ctx, room)
	websocket.GlobalHub.BroadcastToLobby(lobbyID, "BID_PLACED", room)
	s.startTimer(lobbyID)
	return nil
}

func (s *auctionService) startTimer(lobbyID string) {
	s.timerMutex.Lock()
	defer s.timerMutex.Unlock()
	if t, ok := s.timers[lobbyID]; ok {
		t.Stop()
	}
	s.timers[lobbyID] = time.AfterFunc(1*time.Second, func() {
		s.handleTimerTick(lobbyID)
	})
}

func (s *auctionService) stopTimer(lobbyID string) {
	s.timerMutex.Lock()
	defer s.timerMutex.Unlock()
	if t, ok := s.timers[lobbyID]; ok {
		t.Stop()
		delete(s.timers, lobbyID)
	}
}

func (s *auctionService) handleTimerTick(lobbyID string) {
	ctx := context.Background()
	room, err := s.auctionRepo.GetByLobbyID(ctx, lobbyID)
	if err != nil || room == nil || room.Status != "bidding" {
		return
	}
	room.TimerSeconds--
	if room.TimerSeconds <= 0 {
		if room.HighestBidderID.IsZero() {
			s.handlePlayerUnsold(ctx, room)
		} else {
			s.handlePlayerSold(ctx, room)
		}
		return
	}
	s.auctionRepo.Update(ctx, room)
	websocket.GlobalHub.BroadcastToLobby(lobbyID, "TIMER_TICK", map[string]int{"seconds": room.TimerSeconds})
	s.startTimer(lobbyID)
}

func (s *auctionService) handlePlayerSold(ctx context.Context, room *models.AuctionRoom) {
	room.Status = "sold"
	player, _ := s.playerRepo.GetByID(ctx, room.CurrentPlayerID.Hex())
	if player == nil {
		log.Printf("❌ Panic Avoided: Player %s not found in handlePlayerSold", room.CurrentPlayerID.Hex())
		room.Status = "unsold"
		s.auctionRepo.Update(ctx, room)
		return
	}

	lobby, _ := s.lobbyRepo.GetByID(ctx, room.LobbyID.Hex())
	for i, p := range lobby.Participants {
		if p.UserID == room.HighestBidderID {
			lobby.Participants[i].PurseRemaining -= room.CurrentBid
			lobby.Participants[i].Squad = append(lobby.Participants[i].Squad, room.CurrentPlayerID)

			// Update lobby FIRST to ensure persistence of the sale
			if err := s.lobbyRepo.UpdateParticipants(ctx, room.LobbyID.Hex(), lobby.Participants); err != nil {
				log.Printf("❌ Failed to update participants: %v", err)
			}

			record := models.SoldRecord{
				PlayerID:    room.CurrentPlayerID,
				PlayerName:  player.Name,
				PlayerImage: player.ImageUrl,
				PlayerRole:  player.Role,
				WinnerID:    room.HighestBidderID,
				TeamName:    p.TeamName,
				IplTeam:     p.IplTeam,
				SoldPrice:   room.CurrentBid,
				SoldAt:      time.Now(),
			}
			room.SoldPlayers = append(room.SoldPlayers, record)
			break
		}
	}
	s.auctionRepo.Update(ctx, room)
	s.stopTimer(room.LobbyID.Hex()) // Stop timer on sale
	websocket.GlobalHub.BroadcastToLobby(room.LobbyID.Hex(), "PLAYER_SOLD", map[string]interface{}{"room": room, "winnerId": room.HighestBidderID.Hex()})
	websocket.GlobalHub.BroadcastToLobby(room.LobbyID.Hex(), "LOBBY_UPDATED", lobby)
}

func (s *auctionService) handlePlayerUnsold(ctx context.Context, room *models.AuctionRoom) {
	room.Status = "unsold"
	if !room.CurrentPlayerID.IsZero() {
		room.UnsoldPlayers = append(room.UnsoldPlayers, room.CurrentPlayerID)
	}
	s.auctionRepo.Update(ctx, room)
	s.stopTimer(room.LobbyID.Hex()) // Stop timer on unsold
	websocket.GlobalHub.BroadcastToLobby(room.LobbyID.Hex(), "PLAYER_UNSOLD", room)
}

func (s *auctionService) GetAuctionRoom(ctx context.Context, lobbyID string) (interface{}, error) {
	room, err := s.auctionRepo.GetByLobbyID(ctx, lobbyID)
	if err != nil || room == nil {
		// Return a default idle state if auction hasn't started
		return map[string]interface{}{
			"room": map[string]interface{}{
				"status":       "idle",
				"currentBid":   0,
				"timerSeconds": 0,
				"soldPlayers":  []interface{}{},
			},
			"player":   nil,
			"upcoming": []interface{}{},
		}, nil
	}

	upcomingPlayerIDs := []string{}
	for _, id := range room.PlayerQueue {
		upcomingPlayerIDs = append(upcomingPlayerIDs, id.Hex())
	}
	upcomingPlayers, _ := s.playerRepo.GetMany(ctx, upcomingPlayerIDs)

	var player *models.Player
	if !room.CurrentPlayerID.IsZero() {
		player, _ = s.playerRepo.GetByID(ctx, room.CurrentPlayerID.Hex())
	}

	return map[string]interface{}{
		"room":     room,
		"player":   player,
		"upcoming": upcomingPlayers,
	}, nil
}

func (s *auctionService) HandleHostPresenceChange(ctx context.Context, lobbyID string, isPresent bool) error {
	room, err := s.auctionRepo.GetByLobbyID(ctx, lobbyID)
	if err != nil || room == nil {
		return nil // No room, no problem
	}

	room.IsHostPresent = isPresent
	if !isPresent && room.Status == "bidding" {
		// Auto-pause if host leaves during bidding
		room.Status = "paused"
		s.stopTimer(lobbyID)
	}

	if err := s.auctionRepo.Update(ctx, room); err != nil {
		return err
	}

	websocket.GlobalHub.BroadcastToLobby(lobbyID, "AUCTION_STATE_UPDATED", map[string]interface{}{
		"room": room,
	})
	return nil
}
