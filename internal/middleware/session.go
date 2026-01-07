package middleware

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const sessionTTL = 7 * 24 * time.Hour // one week

// Session represents an authenticated user session
type Session struct {
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore defines the session storage interface.
// Implementations could use memory, SQLite, Redis, etc.
type SessionStore interface {
	Create() string
	Get(id string) (Session, bool)
	Delete(id string)
}

// MemorySessionStore is an in-memory implementation of SessionStore.
// Sessions are lost on server restart.
type MemorySessionStore struct {
	mu sync.RWMutex
	m  map[string]Session
}

// NewMemorySessionStore creates a new in-memory session store.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		m: make(map[string]Session),
	}
}

// Create creates a new session and returns its ID.
func (s *MemorySessionStore) Create() string {
	id := uuid.NewString()
	now := time.Now()

	s.mu.Lock()
	s.m[id] = Session{
		CreatedAt: now,
		ExpiresAt: now.Add(sessionTTL),
	}
	s.mu.Unlock()

	return id
}

// Get retrieves a session by ID. Returns false if not found or expired.
func (s *MemorySessionStore) Get(id string) (Session, bool) {
	s.mu.RLock()
	sess, ok := s.m[id]
	s.mu.RUnlock()

	if !ok {
		return Session{}, false
	}

	if time.Now().After(sess.ExpiresAt) {
		// Expired: clean up and treat as missing
		s.mu.Lock()
		delete(s.m, id)
		s.mu.Unlock()
		return Session{}, false
	}

	return sess, true
}

// Delete removes a session by ID.
func (s *MemorySessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.m, id)
	s.mu.Unlock()
}

// Default session store used by package-level functions
var defaultStore SessionStore = NewMemorySessionStore()

// CreateSession creates a new session and returns its ID.
func CreateSession() string {
	return defaultStore.Create()
}

// GetSession retrieves a session by ID. Returns false if not found or expired.
func GetSession(id string) (Session, bool) {
	return defaultStore.Get(id)
}

// DeleteSession removes a session by ID.
func DeleteSession(id string) {
	defaultStore.Delete(id)
}
