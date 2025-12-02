package game

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	RoomMaxPlayers = 2
	RoomWaitTimeout = 10 * time.Second
)

type Player struct {
	ID       string
	UserID   string
	Conn     *websocket.Conn
	Score    int
	Frame    int  // Current pet frame (0-3)
	GameOver bool
	Ready    bool  // Ready after countdown
	mu       sync.Mutex
	writeMu  sync.Mutex // Protects WebSocket writes (must be separate from mu)
}

func NewPlayer(userID string, conn *websocket.Conn) *Player {
	return &Player{
		ID:       uuid.New().String(),
		UserID:   userID,
		Conn:     conn,
		Score:    0,
		Frame:    0,
		GameOver: false,
		Ready:    false,
	}
}

func (p *Player) SetReady(ready bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Ready = ready
}

func (p *Player) IsReady() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Ready
}

func (p *Player) SetFrame(frame int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Frame = frame
}

func (p *Player) GetFrame() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Frame
}

func (p *Player) SetGameOver() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.GameOver = true
}

func (p *Player) IsGameOver() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.GameOver
}

// SendMessage sends a message through the WebSocket connection with mutex protection
// This ensures thread-safe writes to the connection
func (p *Player) SendMessage(msg interface{}) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	return p.Conn.WriteJSON(msg)
}

type RoomState string

const (
	RoomStateWaiting RoomState = "waiting"  // Waiting for second player
	RoomStatePlaying RoomState = "playing"  // Game in progress
	RoomStateEnded   RoomState = "ended"     // Both players game over
)

type SunnySaysRoom struct {
	ID        string
	Players   []*Player
	State     RoomState
	CreatedAt time.Time
	mu        sync.RWMutex
	
	// Game state
	CurrentRound      int
	SunnyFrame        int  // Current Sunny's frame (0-3)
	RoundActive       bool
	ConfusionEnabled  bool
	WaitStartTime     time.Time
	SunnyShowTime     time.Time
	muGame            sync.RWMutex
}

func NewSunnySaysRoom() *SunnySaysRoom {
	return &SunnySaysRoom{
		ID:        uuid.New().String(),
		Players:   make([]*Player, 0, RoomMaxPlayers),
		State:     RoomStateWaiting,
		CreatedAt: time.Now(),
		CurrentRound: 0,
		SunnyFrame: 0,
		RoundActive: false,
		ConfusionEnabled: false,
	}
}

func (r *SunnySaysRoom) AddPlayer(player *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if len(r.Players) >= RoomMaxPlayers {
		return false
	}
	
	r.Players = append(r.Players, player)
	
	if len(r.Players) == RoomMaxPlayers {
		r.State = RoomStatePlaying
	}
	
	return true
}

func (r *SunnySaysRoom) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i, p := range r.Players {
		if p.ID == playerID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			break
		}
	}
	
	if len(r.Players) == 0 {
		r.State = RoomStateEnded
	}
}

func (r *SunnySaysRoom) GetPlayer(playerID string) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, p := range r.Players {
		if p.ID == playerID {
			return p
		}
	}
	return nil
}

func (r *SunnySaysRoom) GetOpponent(playerID string) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, p := range r.Players {
		if p.ID != playerID {
			return p
		}
	}
	return nil
}

func (r *SunnySaysRoom) IsFull() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Players) >= RoomMaxPlayers
}

func (r *SunnySaysRoom) AllPlayersGameOver() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// If no players, all are game over (room should be cleaned up)
	if len(r.Players) == 0 {
		return true
	}
	
	// Check if all remaining players are game over
	for _, p := range r.Players {
		if !p.IsGameOver() {
			return false
		}
	}
	return true
}

func (r *SunnySaysRoom) AllPlayersReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Once game has started (State == Playing), only need active (non-game-over) players to be ready
	// Before game starts, need all players (for initial countdown)
	if r.State == RoomStatePlaying {
		// Game has started - only need active players to be ready
		hasActivePlayers := false
		for _, p := range r.Players {
			if !p.IsGameOver() {
				hasActivePlayers = true
				if !p.IsReady() {
					return false
				}
			}
		}
		return hasActivePlayers // Return true if at least one active player is ready
	} else {
		// Game hasn't started yet - need all players to be ready
		if len(r.Players) < RoomMaxPlayers {
			return false
		}
		
		for _, p := range r.Players {
			if !p.IsReady() {
				return false
			}
		}
		return true
	}
}

func (r *SunnySaysRoom) ResetReady() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, p := range r.Players {
		p.SetReady(false)
	}
}

// Sunny AI Logic (moved from Godot)

func (r *SunnySaysRoom) ChooseRandomSymbol() int {
	// Returns random frame: 1 (heart), 2 (diamond), or 3 (both)
	// Frame 0 (nothing) is not used as Sunny's choice
	return rand.Intn(3) + 1
}

func (r *SunnySaysRoom) ShouldUseConfusion() bool {
	r.muGame.RLock()
	defer r.muGame.RUnlock()
	
	// Check if any player has score >= 3
	for _, p := range r.Players {
		if p.Score >= 3 {
			// Random chance to use confusion (50% chance)
			return rand.Float32() < 0.5
		}
	}
	return false
}

func (r *SunnySaysRoom) GetConfusionSequence() []int {
	// Flash symbols up to 3 times
	flashCount := rand.Intn(3) + 1
	sequence := make([]int, 0, flashCount+1)
	
	for i := 0; i < flashCount; i++ {
		sequence = append(sequence, r.ChooseRandomSymbol())
	}
	
	// Final symbol
	sequence = append(sequence, r.ChooseRandomSymbol())
	
	return sequence
}

func (r *SunnySaysRoom) SetSunnyFrame(frame int) {
	r.muGame.Lock()
	defer r.muGame.Unlock()
	r.SunnyFrame = frame
}

func (r *SunnySaysRoom) GetSunnyFrame() int {
	r.muGame.RLock()
	defer r.muGame.RUnlock()
	return r.SunnyFrame
}

func (r *SunnySaysRoom) SetRoundActive(active bool) {
	r.muGame.Lock()
	defer r.muGame.Unlock()
	r.RoundActive = active
}

func (r *SunnySaysRoom) IsRoundActive() bool {
	r.muGame.RLock()
	defer r.muGame.RUnlock()
	return r.RoundActive
}

func (r *SunnySaysRoom) IncrementRound() {
	r.muGame.Lock()
	defer r.muGame.Unlock()
	r.CurrentRound++
	
	// Enable confusion after any player reaches 3 points
	if !r.ConfusionEnabled {
		for _, p := range r.Players {
			if p.Score >= 3 {
				r.ConfusionEnabled = true
				break
			}
		}
	}
}

