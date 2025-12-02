package game

import (
	"sync"
	"time"
)

type RoomManager struct {
	rooms map[string]*SunnySaysRoom
	mu    sync.RWMutex
}

func NewRoomManager() *RoomManager {
	rm := &RoomManager{
		rooms: make(map[string]*SunnySaysRoom),
	}
	
	// Start cleanup goroutine to remove empty/ended rooms
	go rm.cleanupRoutine()
	
	return rm
}

func (rm *RoomManager) FindOrCreateRoom() *SunnySaysRoom {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// Find an unfilled room
	for _, room := range rm.rooms {
		if !room.IsFull() && room.State == RoomStateWaiting {
			return room
		}
	}
	
	// Create new room if none found
	room := NewSunnySaysRoom()
	rm.rooms[room.ID] = room
	return room
}

func (rm *RoomManager) GetRoom(roomID string) *SunnySaysRoom {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.rooms, roomID)
}

func (rm *RoomManager) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		rm.mu.Lock()
		for id, room := range rm.rooms {
			// Remove rooms that are ended or have been waiting too long (>5 minutes)
			if room.State == RoomStateEnded {
				delete(rm.rooms, id)
			} else if room.State == RoomStateWaiting {
				if time.Since(room.CreatedAt) > 5*time.Minute {
					delete(rm.rooms, id)
				}
			}
		}
		rm.mu.Unlock()
	}
}

