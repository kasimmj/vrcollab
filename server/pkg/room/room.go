// Package room manages multiplayer room state and user lifecycle.
package room

import (
	"errors"
	"sync"
	"time"
)

// User in a room.
type User struct {
	ID       string
	Name     string
	JoinedAt time.Time
	LastSeen time.Time
}

// Room is a shared multiplayer space.
type Room struct {
	ID       string
	Capacity int
	mu       sync.RWMutex
	users    map[string]*User
}

// NewRoom returns a room.
func NewRoom(id string, capacity int) *Room {
	return &Room{ID: id, Capacity: capacity, users: map[string]*User{}}
}

// Join adds a user to the room. Returns error if full or duplicate.
func (r *Room) Join(u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.users) >= r.Capacity {
		return errors.New("room full")
	}
	if _, exists := r.users[u.ID]; exists {
		return errors.New("user already in room")
	}
	u.JoinedAt = time.Now()
	u.LastSeen = u.JoinedAt
	r.users[u.ID] = u
	return nil
}

// Leave removes a user.
func (r *Room) Leave(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, userID)
}

// Users returns a snapshot.
func (r *Room) Users() []*User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	return out
}

// Touch updates a user's LastSeen.
func (r *Room) Touch(userID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return false
	}
	u.LastSeen = time.Now()
	return true
}

// PruneIdle removes users not seen within ttl. Returns count removed.
func (r *Room) PruneIdle(ttl time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-ttl)
	n := 0
	for id, u := range r.users {
		if u.LastSeen.Before(cutoff) {
			delete(r.users, id)
			n++
		}
	}
	return n
}
