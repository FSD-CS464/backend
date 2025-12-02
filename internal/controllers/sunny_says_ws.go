package controllers

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"fsd-backend/internal/auth"
	"fsd-backend/internal/game"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var sunnySaysUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now (adjust for production)
	},
}

const (
	// Message types from client
	MsgTypeJoinRoom    = "join_room"
	MsgTypePlayerInput = "player_input"
	MsgTypeWaitChoice  = "wait_choice" // "wait" or "singleplayer"
	MsgTypeReady       = "ready"       // Client ready after countdown

	// Message types from server
	MsgTypeRoomJoined       = "room_joined"
	MsgTypeRoomFull         = "room_full"
	MsgTypeWaiting          = "waiting"
	MsgTypeGameStart        = "game_start"
	MsgTypeRoundStart       = "round_start"
	MsgTypeSunnyFrame       = "sunny_frame"
	MsgTypeOpponentFrame    = "opponent_frame"
	MsgTypeRoundResult      = "round_result"
	MsgTypeGameOver         = "game_over"
	MsgTypeOpponentGameOver = "opponent_game_over"
	MsgTypeError            = "error"
)

type ClientMessage struct {
	Type   string `json:"type"`
	Frame  int    `json:"frame,omitempty"`
	Choice string `json:"choice,omitempty"` // "wait" or "singleplayer"
}

type ServerMessage struct {
	Type              string `json:"type"`
	RoomID            string `json:"room_id,omitempty"`
	PlayerID          string `json:"player_id,omitempty"`
	OpponentID        string `json:"opponent_id,omitempty"`
	Frame             int    `json:"frame,omitempty"`
	SunnyFrame        int    `json:"sunny_frame,omitempty"`
	Score             int    `json:"score,omitempty"`
	OpponentScore     int    `json:"opponent_score,omitempty"`
	Round             int    `json:"round,omitempty"`
	ConfusionSeq      []int  `json:"confusion_seq,omitempty"`
	WaitTime          int    `json:"wait_time,omitempty"` // milliseconds
	Message           string `json:"message,omitempty"`
	DisplayDurationMs int    `json:"display_duration_ms"` // Duration in milliseconds for client to display this frame (0 = no duration, final frame)
}

type SunnySaysWSHandler struct {
	roomManager *game.RoomManager
	signer      *auth.Signer
	connMutexes sync.Map // Map[*websocket.Conn]*sync.Mutex for thread-safe writes
}

func NewSunnySaysWSHandler(signer *auth.Signer) *SunnySaysWSHandler {
	return &SunnySaysWSHandler{
		roomManager: game.NewRoomManager(),
		signer:      signer,
	}
}

func (h *SunnySaysWSHandler) HandleConnection(c *gin.Context) {
	// Get JWT token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	// Validate JWT
	claims, err := h.signer.Parse(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	userID := claims.UserID

	// Upgrade to WebSocket
	conn, err := sunnySaysUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Find or create room
	room := h.roomManager.FindOrCreateRoom()

	// Create player
	player := game.NewPlayer(userID, conn)

	// Add player to room
	if !room.AddPlayer(player) {
		// Room is full, send error and close
		h.sendMessage(conn, ServerMessage{
			Type:    MsgTypeError,
			Message: "Room is full",
		})
		return
	}

	// Send room joined message
	h.sendMessage(conn, ServerMessage{
		Type:     MsgTypeRoomJoined,
		RoomID:   room.ID,
		PlayerID: player.ID,
	})

	// If room is now full, start game
	if room.IsFull() {
		opponent := room.GetOpponent(player.ID)
		if opponent != nil {
			// Notify both players game is starting
			h.sendMessage(player.Conn, ServerMessage{
				Type:       MsgTypeGameStart,
				OpponentID: opponent.ID,
			})
			h.sendMessage(opponent.Conn, ServerMessage{
				Type:       MsgTypeGameStart,
				OpponentID: player.ID,
			})

			// Start game loop
			go h.runGame(room)
		}
	} else {
		// Waiting for second player
		h.sendMessage(conn, ServerMessage{
			Type: MsgTypeWaiting,
		})

		// Start waiting timeout
		go h.handleWaitingTimeout(room, player)
	}

	// Handle incoming messages
	h.handlePlayerMessages(room, player)
}

func (h *SunnySaysWSHandler) handleWaitingTimeout(room *game.SunnySaysRoom, player *game.Player) {
	timeout := time.NewTimer(game.RoomWaitTimeout)
	defer timeout.Stop()

	<-timeout.C

	// Check if room is still waiting AND game hasn't started
	// Don't send timeout if game has already started
	if room.State == game.RoomStatePlaying {
		// Game has started, don't send timeout message
		return
	}

	// Check if room is still waiting
	isFull := room.IsFull()

	if !isFull {
		// 10 seconds passed, ask player to wait or play singleplayer
		// Only send if player is still in the room and game hasn't started
		if room.GetPlayer(player.ID) != nil {
			h.sendMessage(player.Conn, ServerMessage{
				Type:    MsgTypeWaiting,
				Message: "timeout",
			})
		}
	}
}

func (h *SunnySaysWSHandler) handlePlayerMessages(room *game.SunnySaysRoom, player *game.Player) {
	for {
		var msg ClientMessage
		err := player.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message from player %s: %v", player.ID, err)
			break
		}

		switch msg.Type {
		case MsgTypePlayerInput:
			// Update player's frame
			player.SetFrame(msg.Frame)

			// Broadcast to opponent
			opponent := room.GetOpponent(player.ID)
			if opponent != nil && !opponent.IsGameOver() {
				h.sendMessage(opponent.Conn, ServerMessage{
					Type:  MsgTypeOpponentFrame,
					Frame: msg.Frame,
				})
			}

		case MsgTypeWaitChoice:
			if msg.Choice == "wait" {
				// Wait another 10 seconds - send waiting message to show searching text
				h.sendMessage(player.Conn, ServerMessage{
					Type: MsgTypeWaiting,
				})
				go h.handleWaitingTimeout(room, player)
			} else if msg.Choice == "singleplayer" {
				// Remove player from room and let them play singleplayer
				room.RemovePlayer(player.ID)
				// Client should handle singleplayer mode
				h.sendMessage(player.Conn, ServerMessage{
					Type:    MsgTypeError,
					Message: "singleplayer_mode",
				})
				return
			}

		case MsgTypeReady:
			// Player is ready after countdown
			player.SetReady(true)
			// Check if all players are ready and no round is currently active
			if room.AllPlayersReady() && !room.IsRoundActive() {
				// Both players ready, start round in goroutine
				room.ResetReady() // Reset for next round
				go h.startRound(room)
			}
		}
	}

	// Player disconnected
	// Get opponent BEFORE removing player (so we can notify them)
	var opponent *game.Player
	if room.State == game.RoomStatePlaying {
		opponent = room.GetOpponent(player.ID)
	} else {
		// Game hasn't started - get remaining player if any
		if len(room.Players) > 1 {
			for _, p := range room.Players {
				if p.ID != player.ID {
					opponent = p
					break
				}
			}
		}
	}

	// Remove player from room
	room.RemovePlayer(player.ID)

	// Notify opponent if game has started and they're still active
	if room.State == game.RoomStatePlaying && opponent != nil && !opponent.IsGameOver() {
		// Opponent is still playing - notify them that opponent disconnected
		// They can continue playing solo - connection stays open
		// Send message, but don't close connection even if there's an error
		if err := h.sendMessageSafe(opponent.Conn, ServerMessage{
			Type: MsgTypeOpponentGameOver,
		}); err != nil {
			log.Printf("Failed to notify opponent of disconnect: %v", err)
			// Don't close connection - let them continue playing
		}

		// If a round is active and only one player remains, mark round as inactive
		// This allows the remaining player to continue playing
		if room.IsRoundActive() && len(room.Players) == 1 {
			room.SetRoundActive(false)
			// Reset ready status so the remaining player can send ready to start next round
			room.ResetReady()
		}
	} else if room.State != game.RoomStatePlaying && opponent != nil {
		// Game hasn't started yet - notify opponent
		h.sendMessage(opponent.Conn, ServerMessage{
			Type:    MsgTypeError,
			Message: "opponent_disconnected",
		})
	}

	// Clean up room only if empty
	// Note: We don't check AllPlayersGameOver() here because game-over players
	// can disconnect and leave, allowing them to join new games
	if len(room.Players) == 0 {
		h.roomManager.RemoveRoom(room.ID)
	}
}

func (h *SunnySaysWSHandler) runGame(room *game.SunnySaysRoom) {
	// Wait for both players to be ready (after countdown)
	// This is handled by the ready message system
	// The first round will start when both players send ready

	// Game loop - wait for ready messages to trigger rounds
	for {
		// Check if all players game over
		if room.AllPlayersGameOver() {
			// Notify both players
			for _, p := range room.Players {
				h.sendMessage(p.Conn, ServerMessage{
					Type: MsgTypeGameOver,
				})
			}
			room.State = game.RoomStateEnded
			h.roomManager.RemoveRoom(room.ID)
			return
		}

		// Wait a bit to avoid busy loop
		time.Sleep(100 * time.Millisecond)
	}
}

func (h *SunnySaysWSHandler) startRound(room *game.SunnySaysRoom) {
	// Prevent multiple rounds from running simultaneously
	if room.IsRoundActive() {
		log.Printf("Round already active, skipping startRound")
		return
	}

	// Check if all players are game over - if so, don't start a new round
	if room.AllPlayersGameOver() {
		log.Printf("All players game over, not starting new round")
		return
	}

	// Check if there are any active (non-game-over) players
	hasActivePlayers := false
	for _, p := range room.Players {
		if !p.IsGameOver() {
			hasActivePlayers = true
			break
		}
	}
	if !hasActivePlayers {
		log.Printf("No active players, not starting new round")
		return
	}

	// Game constants
	const (
		minWaitTime  = 500 * time.Millisecond
		maxWaitTime  = 3000 * time.Millisecond
		matchTimeout = 1000 * time.Millisecond // 1 second for players to match
	)

	// Start new round
	room.IncrementRound()
	room.SetRoundActive(true) // Mark round as active immediately

	// Reset active players to frame 0 and send round start message
	for _, p := range room.Players {
		if !p.IsGameOver() {
			p.SetFrame(0)
			h.sendMessage(p.Conn, ServerMessage{
				Type:  MsgTypeRoundStart,
				Round: room.CurrentRound,
			})
		}
	}

	// Wait random time before Sunny shows symbol (0.5 to 3 seconds)
	waitMs := 500 + rand.Int63n(2500) // 500ms to 3000ms
	time.Sleep(time.Duration(waitMs) * time.Millisecond)

	// Check if confusion should be used
	var sunnyFrame int
	useConfusion := room.ShouldUseConfusion()

	if useConfusion {
		// Run confusion sequence - flash 1-3 times, then show final symbol
		flashCount := rand.Intn(3) + 1 // 1 to 3 flashes

		// Flash symbols
		for i := 0; i < flashCount; i++ {
			// Choose random symbol for this flash
			flashFrame := room.ChooseRandomSymbol()
			room.SetSunnyFrame(flashFrame)

			// Send flash frame with 300ms duration to all players
			duration300 := 300
			for _, p := range room.Players {
				if !p.IsGameOver() {
					h.sendMessage(p.Conn, ServerMessage{
						Type:              MsgTypeSunnyFrame,
						SunnyFrame:        flashFrame,
						DisplayDurationMs: duration300, // Client should display this for 300ms
					})
					// Small delay to ensure messages are properly spaced
					time.Sleep(10 * time.Millisecond)
				}
			}

			// Server waits for the flash duration (keeps server in sync)
			time.Sleep(300 * time.Millisecond)

			// If not the last flash, wait random time, reset to idle, then wait briefly
			if i < flashCount-1 {
				// Wait random time between flashes (0.3 to 1.0 seconds)
				flashWaitMs := 300 + rand.Int63n(700)  // 300ms to 1000ms
				totalIdleDuration := flashWaitMs + 100 // flashWaitMs + brief pause

				// Reset to idle (frame 0) with combined duration
				room.SetSunnyFrame(0)
				idleDuration := int(totalIdleDuration)
				for _, p := range room.Players {
					if !p.IsGameOver() {
						h.sendMessage(p.Conn, ServerMessage{
							Type:              MsgTypeSunnyFrame,
							SunnyFrame:        0,
							DisplayDurationMs: idleDuration, // Combined wait time
						})
						// Small delay to ensure messages are properly spaced
						time.Sleep(10 * time.Millisecond)
					}
				}

				// Server waits for the total idle duration (keeps server in sync)
				time.Sleep(time.Duration(totalIdleDuration) * time.Millisecond)
			}
		}

		// Show final symbol (after all flashes) - no duration needed, this is the final symbol for the round
		sunnyFrame = room.ChooseRandomSymbol()
		room.SetSunnyFrame(sunnyFrame)
		finalDuration := 0 // 0 means no duration, final frame
		for _, p := range room.Players {
			if !p.IsGameOver() {
				h.sendMessage(p.Conn, ServerMessage{
					Type:              MsgTypeSunnyFrame,
					SunnyFrame:        sunnyFrame,
					DisplayDurationMs: finalDuration, // 0 = final frame, no duration
				})
				// Small delay to ensure messages are properly spaced
				time.Sleep(10 * time.Millisecond)
			}
		}
	} else {
		// Normal round - Sunny shows symbol immediately (no confusion)
		sunnyFrame = room.ChooseRandomSymbol()
		room.SetSunnyFrame(sunnyFrame)

		// Broadcast Sunny frame (no duration, this is the final symbol for the round)
		normalDuration := 0 // 0 means no duration, final frame
		for _, p := range room.Players {
			if !p.IsGameOver() {
				h.sendMessage(p.Conn, ServerMessage{
					Type:              MsgTypeSunnyFrame,
					SunnyFrame:        sunnyFrame,
					DisplayDurationMs: normalDuration, // 0 = final frame, no duration
				})
			}
		}
	}

	// Round is already active (set at start of function)
	// Wait for match timeout
	time.Sleep(matchTimeout)

	// Check results for each player
	for _, p := range room.Players {
		if p.IsGameOver() {
			continue
		}

		playerFrame := p.GetFrame()
		matched := playerFrame == sunnyFrame && playerFrame != 0

		if matched {
			// Match successful
			p.Score++
			h.sendMessage(p.Conn, ServerMessage{
				Type:  MsgTypeRoundResult,
				Score: p.Score,
				Frame: playerFrame,
			})
		} else {
			// Game over for this player
			p.SetGameOver()
			h.sendMessage(p.Conn, ServerMessage{
				Type:  MsgTypeGameOver,
				Score: p.Score,
			})

			// Notify opponent
			opponent := room.GetOpponent(p.ID)
			if opponent != nil {
				h.sendMessage(opponent.Conn, ServerMessage{
					Type: MsgTypeOpponentGameOver,
				})
			}
		}
	}

	// Wait before next round
	time.Sleep(500 * time.Millisecond)

	// Mark round as inactive - ready for next round
	room.SetRoundActive(false)

	// Reset ready status for next round
	room.ResetReady()

	// Don't recursively call startRound - wait for ready messages to trigger next round
	// The ready message handler will start the next round when both players are ready
}

func (h *SunnySaysWSHandler) sendMessage(conn *websocket.Conn, msg ServerMessage) {
	// Get or create mutex for this connection
	mutexInterface, _ := h.connMutexes.LoadOrStore(conn, &sync.Mutex{})
	writeMu := mutexInterface.(*sync.Mutex)

	writeMu.Lock()
	defer writeMu.Unlock()

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// sendMessageSafe sends a message and returns an error if it fails
// This is useful when we want to handle errors without closing connections
func (h *SunnySaysWSHandler) sendMessageSafe(conn *websocket.Conn, msg ServerMessage) error {
	// Get or create mutex for this connection
	mutexInterface, _ := h.connMutexes.LoadOrStore(conn, &sync.Mutex{})
	writeMu := mutexInterface.(*sync.Mutex)

	writeMu.Lock()
	defer writeMu.Unlock()

	return conn.WriteJSON(msg)
}
