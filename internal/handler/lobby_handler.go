package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	gwebsocket "github.com/gorilla/websocket"

	"bidxi/internal/dto"
	"bidxi/internal/service"
	"bidxi/internal/utils"
	"bidxi/internal/websocket"
)

// LobbyHandler handles lobby-related HTTP requests.
type LobbyHandler struct {
	lobbyService service.LobbyService
}

// NewLobbyHandler creates a new LobbyHandler.
func NewLobbyHandler(lobbyService service.LobbyService) *LobbyHandler {
	return &LobbyHandler{lobbyService: lobbyService}
}

// RegisterRoutes registers lobby-related routes on the protected group.
func (h *LobbyHandler) RegisterRoutes(rg *gin.RouterGroup) {
	lobbies := rg.Group("/lobbies")
	{
		lobbies.POST("", h.CreateLobby)
		lobbies.GET("", h.GetMyLobbies)
		lobbies.GET(":id", h.GetLobby)
		lobbies.GET(":id/ws", h.ConnectWebSocket)
		lobbies.POST(":id/join", h.JoinLobby)
		lobbies.DELETE(":id/participants/:uid", h.KickParticipant)
		lobbies.POST(":id/start", h.StartAuction)
		lobbies.GET(":id/auction", h.GetAuctionRoom)
		lobbies.GET(":id/squads", h.GetLobbySquads)
		lobbies.POST(":id/bid", h.PlaceBid)
		lobbies.POST(":id/auction/toggle", h.ToggleAuction)
		lobbies.GET(":id/all-players", h.GetAllPlayers)
	}
	// NOTE: join-by-code is intentionally at /api/join-by-code (not /api/lobbies/join-by-code)
	// because Gin's httprouter cannot mix static paths and wildcard params at the same
	// segment level for the same HTTP method (POST /lobbies/:id vs POST /lobbies/join-by-code).
	rg.POST("/join-by-code", h.JoinByCode)
}

func (h *LobbyHandler) CreateLobby(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CreateLobbyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request body", err)
		return
	}

	lobby, status, err := h.lobbyService.CreateLobby(c.Request.Context(), userID.(string), &req)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "Lobby created successfully", lobby)
}

func (h *LobbyHandler) GetLobby(c *gin.Context) {
	lobbyID := c.Param("id")
	lobby, status, err := h.lobbyService.GetLobby(c.Request.Context(), lobbyID)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "Lobby fetched", lobby)
}

// ConnectWebSocket handles upgrading HTTP to WebSocket for a given lobby
func (h *LobbyHandler) ConnectWebSocket(c *gin.Context) {
	lobbyID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var upgrader = gwebsocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // In production, verify origins!
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// upgrader.Upgrade already replies to the client
		return
	}

	// Add client to hub
	client := &websocket.Client{
		Hub:     websocket.GlobalHub,
		Conn:    conn,
		LobbyID: lobbyID,
		UserID:  userID.(string),
	}
	websocket.GlobalHub.AddClient(client)

	// Host presence detection
	lobby, _, _ := h.lobbyService.GetLobby(c.Request.Context(), lobbyID)
	if lobby != nil && lobby.AdminID == userID.(string) {
		h.lobbyService.HandleHostPresenceChange(c.Request.Context(), lobbyID, true)
		defer h.lobbyService.HandleHostPresenceChange(context.Background(), lobbyID, false)
	}

	defer websocket.GlobalHub.RemoveClient(client)

	// Keep connection alive until client closes or error
	defer func() {
		websocket.GlobalHub.RemoveClient(client)
		client.Conn.Close()
	}()

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *LobbyHandler) GetMyLobbies(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, status, err := h.lobbyService.GetMyLobbies(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "User lobbies retrieved successfully", res)
}

func (h *LobbyHandler) JoinLobby(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	lobbyID := c.Param("id")
	var req dto.JoinLobbyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request body", err)
		return
	}

	lobby, status, err := h.lobbyService.JoinLobby(c.Request.Context(), userID.(string), lobbyID, &req)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "Joined lobby", lobby)
}

// JoinByCode allows joining a lobby using only an invite code (no lobby ID needed).
func (h *LobbyHandler) JoinByCode(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.JoinByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request body", err)
		return
	}

	lobby, status, err := h.lobbyService.JoinByCode(c.Request.Context(), userID.(string), &req)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "Joined lobby", lobby)
}

func (h *LobbyHandler) KickParticipant(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	lobbyID := c.Param("id")
	participantID := c.Param("uid")

	status, err := h.lobbyService.KickParticipant(c.Request.Context(), userID.(string), lobbyID, participantID)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "Participant kicked", nil)
}

func (h *LobbyHandler) StartAuction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	lobbyID := c.Param("id")
	status, err := h.lobbyService.StartAuction(c.Request.Context(), userID.(string), lobbyID)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "Auction started successfully", nil)
}

func (h *LobbyHandler) PlaceBid(c *gin.Context) {
	lobbyID := c.Param("id")
	userID, _ := c.Get("userID")

	var req struct {
		Amount int `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid bid amount", err)
		return
	}

	if err := h.lobbyService.PlaceBid(c.Request.Context(), lobbyID, userID.(string), req.Amount); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Bid placed successfully", nil)
}
func (h *LobbyHandler) GetAuctionRoom(c *gin.Context) {
	lobbyID := c.Param("id")
	data, err := h.lobbyService.GetAuctionRoom(c.Request.Context(), lobbyID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Auction room fetched", data)
}

func (h *LobbyHandler) GetLobbySquads(c *gin.Context) {
	lobbyID := c.Param("id")
	data, err := h.lobbyService.GetLobbySquads(c.Request.Context(), lobbyID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Squads fetched", data)
}

func (h *LobbyHandler) GetAllPlayers(c *gin.Context) {
	data, err := h.lobbyService.GetAllPlayers(c.Request.Context())
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, "Players fetched", data)
}

func (h *LobbyHandler) ToggleAuction(c *gin.Context) {
	lobbyID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Action string `json:"action"` // start, pause, resume, next
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid action")
		return
	}

	// Verify Admin
	lobby, status, err := h.lobbyService.GetLobby(c.Request.Context(), lobbyID)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	if lobby.AdminID != userID.(string) {
		utils.RespondError(c, http.StatusForbidden, "Only admin can control auction")
		return
	}

	err = h.lobbyService.ToggleAuction(c.Request.Context(), lobbyID, req.Action)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Auction state updated", nil)
}
