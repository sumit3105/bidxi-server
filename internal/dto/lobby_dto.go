package dto

// CreateLobbyRequest represents payload to create a new lobby.
type CreateLobbyRequest struct {
	Name       string `json:"name" binding:"required,min=3,max=50"`
	MaxPlayers int    `json:"maxPlayers" binding:"required,min=2,max=10"`
	Purse      int    `json:"purse" binding:"required,min=50,max=1000"`
	TeamName   string `json:"teamName" binding:"required,min=3,max=30"`
	IplTeam    string `json:"iplTeam" binding:"required"`
}

// JoinLobbyRequest represents payload for joining a lobby by its ID.
type JoinLobbyRequest struct {
	InviteCode string `json:"inviteCode" binding:"required,len=6"`
	TeamName   string `json:"teamName" binding:"required,min=3,max=30"`
	IplTeam    string `json:"iplTeam" binding:"required"`
}

// JoinByCodeRequest represents payload for joining a lobby using only an invite code.
type JoinByCodeRequest struct {
	InviteCode string `json:"inviteCode" binding:"required,len=6"`
	TeamName   string `json:"teamName" binding:"required,min=3,max=30"`
	IplTeam    string `json:"iplTeam" binding:"required"`
}

// ParticipantDTO is returned in lobby responses.
type ParticipantDTO struct {
	UserID         string  `json:"userId"`
	Username       string  `json:"username"`
	TeamName       string  `json:"teamName"`
	IplTeam        string  `json:"iplTeam"`
	PurseRemaining float64 `json:"purseRemaining"`
}

// LobbyResponse is the public representation of a lobby.
type LobbyResponse struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	AdminID      string           `json:"adminId"`
	InviteCode   string           `json:"inviteCode"`
	MaxPlayers   int              `json:"maxPlayers"`
	Purse        float64          `json:"purse"`
	Status       string           `json:"status"`
	Participants []ParticipantDTO `json:"participants"`
	CreatedAt    string           `json:"createdAt"`
}
