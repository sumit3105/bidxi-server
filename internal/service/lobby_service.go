package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bidxi/internal/dto"
	"bidxi/internal/models"
	"bidxi/internal/repository"
	"bidxi/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LobbyService provides business logic for lobbies.
type LobbyService interface {
	CreateLobby(ctx context.Context, userID string, req *dto.CreateLobbyRequest) (*dto.LobbyResponse, int, error)
	GetLobby(ctx context.Context, lobbyID string) (*dto.LobbyResponse, int, error)
	GetMyLobbies(ctx context.Context, userID string) ([]*dto.LobbyResponse, int, error)
	JoinLobby(ctx context.Context, userID string, lobbyID string, req *dto.JoinLobbyRequest) (*dto.LobbyResponse, int, error)
	JoinByCode(ctx context.Context, userID string, req *dto.JoinByCodeRequest) (*dto.LobbyResponse, int, error)
	KickParticipant(ctx context.Context, userID string, lobbyID string, participantID string) (int, error)
	StartAuction(ctx context.Context, userID string, lobbyID string) (int, error)
	ToggleAuction(ctx context.Context, lobbyID string, action string) error
	GetAuctionRoom(ctx context.Context, lobbyID string) (interface{}, error)
	PlaceBid(ctx context.Context, lobbyID string, userID string, amount int) error
	HandleHostPresenceChange(ctx context.Context, lobbyID string, isPresent bool) error
	GetLobbySquads(ctx context.Context, lobbyID string) (interface{}, error)
	GetAllPlayers(ctx context.Context) (interface{}, error)
}

type lobbyService struct {
	lobbyRepo      repository.LobbyRepository
	userRepo       repository.UserRepository
	auctionService AuctionService
}

// NewLobbyService initializes a LobbyService implementation.
func NewLobbyService(lobbyRepo repository.LobbyRepository, userRepo repository.UserRepository, auctionService AuctionService) LobbyService {
	return &lobbyService{lobbyRepo: lobbyRepo, userRepo: userRepo, auctionService: auctionService}
}

func (s *lobbyService) CreateLobby(ctx context.Context, userID string, req *dto.CreateLobbyRequest) (*dto.LobbyResponse, int, error) {
	if req.MaxPlayers < 2 || req.MaxPlayers > 10 {
		return nil, http.StatusBadRequest, fmt.Errorf("maxPlayers must be between 2 and 10")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID")
	}

	inviteCode, err := s.buildUniqueInviteCode(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	username := s.resolveUsername(ctx, userID)

	// Admin is automatically the first participant
	adminParticipant := models.Participant{
		UserID:         userObjID,
		Username:       username,
		TeamName:       strings.TrimSpace(req.TeamName),
		IplTeam:        req.IplTeam,
		PurseRemaining: req.Purse * 100, // Store in Lakhs
		JoinedAt:       time.Now(),
	}

	lobby := &models.Lobby{
		Name:         strings.TrimSpace(req.Name),
		AdminID:      userObjID,
		InviteCode:   inviteCode,
		MaxPlayers:   req.MaxPlayers,
		Purse:        req.Purse * 100, // Store in Lakhs
		Status:       "waiting",
		Participants: []models.Participant{adminParticipant},
	}

	if err := s.lobbyRepo.Create(ctx, lobby); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create lobby: %w", err)
	}

	return toLobbyDTO(lobby), http.StatusCreated, nil
}

func (s *lobbyService) GetLobby(ctx context.Context, lobbyID string) (*dto.LobbyResponse, int, error) {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch lobby: %w", err)
	}
	if lobby == nil {
		return nil, http.StatusNotFound, fmt.Errorf("lobby not found")
	}
	return toLobbyDTO(lobby), http.StatusOK, nil
}

func (s *lobbyService) GetMyLobbies(ctx context.Context, userID string) ([]*dto.LobbyResponse, int, error) {
	lobbies, err := s.lobbyRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get user lobbies: %w", err)
	}

	var dtoList []*dto.LobbyResponse
	for _, l := range lobbies {
		dtoList = append(dtoList, toLobbyDTO(l))
	}
	if dtoList == nil {
		dtoList = []*dto.LobbyResponse{}
	}

	return dtoList, http.StatusOK, nil
}

func (s *lobbyService) JoinLobby(ctx context.Context, userID string, lobbyID string, req *dto.JoinLobbyRequest) (*dto.LobbyResponse, int, error) {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if lobby == nil {
		return nil, http.StatusNotFound, fmt.Errorf("lobby not found")
	}
	if !strings.EqualFold(lobby.InviteCode, req.InviteCode) {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid invite code")
	}
	if lobby.Status != "waiting" {
		return nil, http.StatusConflict, fmt.Errorf("lobby already started")
	}
	if len(lobby.Participants) >= lobby.MaxPlayers {
		return nil, http.StatusConflict, fmt.Errorf("lobby is full")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID")
	}

	for _, p := range lobby.Participants {
		if p.UserID == userObjID {
			return nil, http.StatusConflict, fmt.Errorf("already joined")
		}
		if strings.EqualFold(p.TeamName, req.TeamName) {
			return nil, http.StatusConflict, fmt.Errorf("team name already taken")
		}
	}

	username := s.resolveUsername(ctx, userID)

	participant := models.Participant{
		UserID:         userObjID,
		Username:       username,
		TeamName:       strings.TrimSpace(req.TeamName),
		IplTeam:        req.IplTeam,
		PurseRemaining: lobby.Purse, // Already in Lakhs
		JoinedAt:       time.Now(),
	}

	if err := s.lobbyRepo.AddParticipant(ctx, lobbyID, participant); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	lobby, err = s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil || lobby == nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch lobby after join")
	}

	dtoLobby := toLobbyDTO(lobby)
	websocket.GlobalHub.BroadcastToLobby(lobbyID, "LOBBY_UPDATED", dtoLobby)

	return dtoLobby, http.StatusOK, nil
}

// JoinByCode allows a user to join a lobby using only the invite code (no lobby ID required).
func (s *lobbyService) JoinByCode(ctx context.Context, userID string, req *dto.JoinByCodeRequest) (*dto.LobbyResponse, int, error) {
	code := strings.ToUpper(strings.TrimSpace(req.InviteCode))
	lobby, err := s.lobbyRepo.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to look up invite code: %w", err)
	}
	if lobby == nil {
		return nil, http.StatusNotFound, fmt.Errorf("no lobby found with this invite code")
	}
	if lobby.Status != "waiting" {
		return nil, http.StatusConflict, fmt.Errorf("lobby already started or completed")
	}
	if len(lobby.Participants) >= lobby.MaxPlayers {
		return nil, http.StatusConflict, fmt.Errorf("lobby is full")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid user ID")
	}

	for _, p := range lobby.Participants {
		if p.UserID == userObjID {
			// Already a participant — return current lobby state (idempotent)
			return toLobbyDTO(lobby), http.StatusOK, nil
		}
		if strings.EqualFold(p.TeamName, req.TeamName) {
			return nil, http.StatusConflict, fmt.Errorf("team name already taken in this lobby")
		}
	}

	username := s.resolveUsername(ctx, userID)

	participant := models.Participant{
		UserID:         userObjID,
		Username:       username,
		TeamName:       strings.TrimSpace(req.TeamName),
		IplTeam:        req.IplTeam,
		PurseRemaining: lobby.Purse,
		JoinedAt:       time.Now(),
	}

	lobbyID := lobby.ID.Hex()
	if err := s.lobbyRepo.AddParticipant(ctx, lobbyID, participant); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to add participant: %w", err)
	}

	lobby, err = s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil || lobby == nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch lobby after join")
	}

	dtoLobby := toLobbyDTO(lobby)
	websocket.GlobalHub.BroadcastToLobby(lobbyID, "LOBBY_UPDATED", dtoLobby)

	return dtoLobby, http.StatusOK, nil
}

func (s *lobbyService) KickParticipant(ctx context.Context, userID string, lobbyID string, participantID string) (int, error) {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if lobby == nil {
		return http.StatusNotFound, fmt.Errorf("lobby not found")
	}

	if lobby.AdminID.Hex() != userID {
		return http.StatusForbidden, fmt.Errorf("only lobby admin can kick participants")
	}

	// Send KICKED event to the participant BEFORE removing them
	websocket.GlobalHub.SendToUser(lobbyID, participantID, "KICKED", map[string]string{
		"reason": "You have been removed from the lobby by the host.",
	})

	if err := s.lobbyRepo.RemoveParticipant(ctx, lobbyID, participantID); err != nil {
		return http.StatusInternalServerError, err
	}

	// Fetch updated lobby to broadcast to remaining participants
	updatedLobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err == nil && updatedLobby != nil {
		websocket.GlobalHub.BroadcastToLobby(lobbyID, "LOBBY_UPDATED", toLobbyDTO(updatedLobby))
	}

	return http.StatusOK, nil
}

func (s *lobbyService) StartAuction(ctx context.Context, userID string, lobbyID string) (int, error) {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if lobby == nil {
		return http.StatusNotFound, fmt.Errorf("lobby not found")
	}
	if lobby.AdminID.Hex() != userID {
		return http.StatusForbidden, fmt.Errorf("only lobby admin can start the auction")
	}

	if len(lobby.Participants) < 2 {
		return http.StatusBadRequest, fmt.Errorf("need at least 2 participants to start the auction")
	}

	if lobby.Status != "waiting" {
		return http.StatusConflict, fmt.Errorf("auction already started or completed")
	}
	if len(lobby.Participants) < 2 {
		return http.StatusBadRequest, fmt.Errorf("at least 2 participants required to start auction")
	}

	if err := s.lobbyRepo.UpdateStatus(ctx, lobbyID, "active"); err != nil {
		return http.StatusInternalServerError, err
	}

	updatedLobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err == nil && updatedLobby != nil {
		websocket.GlobalHub.BroadcastToLobby(lobbyID, "LOBBY_UPDATED", toLobbyDTO(updatedLobby))
	}

	// Initialize the Auction Room and start the first round
	_, err = s.auctionService.InitializeAuction(ctx, lobbyID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to initialize auction room: %w", err)
	}

	return http.StatusOK, nil
}

func (s *lobbyService) PlaceBid(ctx context.Context, lobbyID string, userID string, amount int) error {
	return s.auctionService.PlaceBid(ctx, lobbyID, userID, amount)
}

// resolveUsername fetches the username for a given userID; falls back to empty string on error.
func (s *lobbyService) resolveUsername(ctx context.Context, userID string) string {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ""
	}
	return user.Username
}

func (s *lobbyService) buildUniqueInviteCode(ctx context.Context) (string, error) {
	for i := 0; i < 6; i++ {
		code, err := generateInviteCode(6)
		if err != nil {
			return "", err
		}

		exists, err := s.lobbyRepo.GetByInviteCode(ctx, code)
		if err != nil {
			return "", err
		}
		if exists == nil {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique invite code")
}

func generateInviteCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"[bytes[i]%36]
	}
	return strings.ToUpper(string(bytes)), nil
}

func toLobbyDTO(lobby *models.Lobby) *dto.LobbyResponse {
	participants := make([]dto.ParticipantDTO, 0, len(lobby.Participants))
	for _, p := range lobby.Participants {
		participants = append(participants, dto.ParticipantDTO{
			UserID:         p.UserID.Hex(),
			Username:       p.Username,
			TeamName:       p.TeamName,
			IplTeam:        p.IplTeam,
			PurseRemaining: float64(p.PurseRemaining) / 100.0,
		})
	}

	return &dto.LobbyResponse{
		ID:           lobby.ID.Hex(),
		Name:         lobby.Name,
		AdminID:      lobby.AdminID.Hex(),
		InviteCode:   lobby.InviteCode,
		MaxPlayers:   lobby.MaxPlayers,
		Purse:        float64(lobby.Purse) / 100.0,
		Status:       lobby.Status,
		Participants: participants,
		CreatedAt:    lobby.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (s *lobbyService) GetAuctionRoom(ctx context.Context, lobbyID string) (interface{}, error) {
	return s.auctionService.GetAuctionRoom(ctx, lobbyID)
}

func (s *lobbyService) ToggleAuction(ctx context.Context, lobbyID string, action string) error {
	return s.auctionService.ToggleAuctionState(ctx, lobbyID, action)
}

func (s *lobbyService) HandleHostPresenceChange(ctx context.Context, lobbyID string, isPresent bool) error {
	return s.auctionService.HandleHostPresenceChange(ctx, lobbyID, isPresent)
}

func (s *lobbyService) GetLobbySquads(ctx context.Context, lobbyID string) (interface{}, error) {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return nil, err
	}

	room, _ := s.auctionService.(*auctionService).auctionRepo.GetByLobbyID(ctx, lobbyID)
	soldPrices := make(map[string]int)
	if room != nil {
		for _, sp := range room.SoldPlayers {
			soldPrices[sp.PlayerID.Hex()] = sp.SoldPrice
		}
	}

	result := []map[string]interface{}{}
	for _, p := range lobby.Participants {
		playerIDs := []string{}
		for _, id := range p.Squad {
			playerIDs = append(playerIDs, id.Hex())
		}
		players, _ := s.auctionService.(*auctionService).playerRepo.GetMany(ctx, playerIDs)

		enrichedPlayers := []map[string]interface{}{}
		for _, pl := range players {
			price := pl.BasePrice
			if sp, ok := soldPrices[pl.ID.Hex()]; ok {
				price = sp
			}
			enrichedPlayers = append(enrichedPlayers, map[string]interface{}{
				"id":          pl.ID.Hex(),
				"name":        pl.Name,
				"role":        pl.Role,
				"nationality": pl.Nationality,
				"category":    pl.Category,
				"basePrice":   pl.BasePrice,
				"imageUrl":    pl.ImageUrl,
				"soldPrice":   price,
			})
		}

		result = append(result, map[string]interface{}{
			"userId":         p.UserID.Hex(),
			"teamName":       p.TeamName,
			"iplTeam":        p.IplTeam,
			"purseRemaining": float64(p.PurseRemaining) / 100.0,
			"squad":          enrichedPlayers,
		})
	}
	return result, nil
}
func (s *lobbyService) GetAllPlayers(ctx context.Context) (interface{}, error) {
	players, err := s.auctionService.(*auctionService).playerRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return players, nil
}
