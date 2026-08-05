package model

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type SessionState string

const (
	StateIdle                  SessionState = "idle"
	StateAwaitingPhone         SessionState = "awaiting_phone"
	StateLoggingIn             SessionState = "logging_in"
	StateAwaitingOTP           SessionState = "awaiting_otp"
	StateAwaitingOfferID       SessionState = "awaiting_offer_id"
	StateAwaitingAutoInt       SessionState = "awaiting_auto_int"
	StateAwaitingAutoThreshold SessionState = "awaiting_auto_threshold"
	StateAwaitingAutoOffer     SessionState = "awaiting_auto_offer"
	StateLoggedIn              SessionState = "logged_in"
)

type Session struct {
	Phone         string       `json:"phone"`
	FullPhone     string       `json:"full_phone"`
	AccessAuth    string       `json:"access_auth"`
	Authorization string       `json:"authorization"`
	Hash          string       `json:"hash"`
	XDevice       string       `json:"x_device"`
	WebAppVersion string       `json:"web_app_version"`
	State         SessionState `json:"state"`
	LastLoginAt   time.Time    `json:"last_login_at"`

	PendingAuthId     string `json:"pending_auth_id,omitempty"`
	PendingAmlbCookie string `json:"pending_amlbcookie,omitempty"`
	PendingOfferID    string `json:"pending_offer_id,omitempty"`
	PendingPayment    string `json:"pending_payment,omitempty"`

	AutoBuyInterval  int    `json:"auto_buy_interval"`
	AutoBuyThreshold int    `json:"auto_buy_threshold"`
	AutoBuyPackage   string `json:"auto_buy_package"`
	AutoBuyPayment   string `json:"auto_buy_payment"`
	AutoBuyActive    bool   `json:"auto_buy_active"`
	AutoBuyOrderID   string `json:"auto_buy_order_id,omitempty"`
}

func (s *Session) IsLoggedIn() bool {
	return s.State == StateLoggedIn && s.AccessAuth != ""
}

func (s *Session) IsAwaiting() bool {
	return s.State == StateAwaitingPhone || s.State == StateAwaitingOTP || s.State == StateLoggingIn
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session // Key adalah FullPhone (contoh: 62812xxx)
	active   map[int64]string    // Key adalah Telegram UserID, Value adalah FullPhone
	filename string
}

func NewSessionManager(filename string) *SessionManager {
	if filename != "" {
		dir := filepath.Dir(filename)
		os.MkdirAll(dir, 0755)
	}

	sm := &SessionManager{
		sessions: make(map[string]*Session),
		active:   make(map[int64]string),
		filename: filename,
	}
	if filename != "" {
		sm.LoadFromFile()
	}
	return sm
}

func (sm *SessionManager) Get(phone string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[phone]
}

func (sm *SessionManager) Set(phone string, session *Session) {
	sm.mu.Lock()
	sm.sessions[phone] = session
	sm.mu.Unlock()
	sm.SaveToFile()
}

func (sm *SessionManager) Delete(phone string) {
	sm.mu.Lock()
	delete(sm.sessions, phone)
	for userID, activePhone := range sm.active {
		if activePhone == phone {
			delete(sm.active, userID)
		}
	}
	sm.mu.Unlock()
	sm.SaveToFile()
}

func (sm *SessionManager) List() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result := make([]*Session, 0, len(sm.sessions))
	for _, v := range sm.sessions {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FullPhone < result[j].FullPhone
	})
	return result
}

func (sm *SessionManager) SetActive(userID int64, phone string) {
	sm.mu.Lock()
	sm.active[userID] = phone
	sm.mu.Unlock()
	sm.SaveToFile()
}

func (sm *SessionManager) GetActive(userID int64) *Session {
	sm.mu.RLock()
	phone, ok := sm.active[userID]
	sm.mu.RUnlock()
	if !ok || phone == "" {
		return nil
	}
	return sm.Get(phone)
}

type sessionData struct {
	Sessions map[string]*Session `json:"sessions"`
	Active   map[int64]string    `json:"active,omitempty"`
}

func (sm *SessionManager) SaveToFile() {
	if sm.filename == "" {
		return
	}

	sm.mu.RLock()
	data := sessionData{
		Sessions: sm.sessions,
		Active:   sm.active,
	}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	sm.mu.RUnlock()

	if err != nil {
		log.Printf("❌ Failed to marshal sessions: %v", err)
		return
	}

	if err := os.WriteFile(sm.filename, jsonData, 0600); err != nil {
		log.Printf("❌ Failed to save sessions to %s: %v", sm.filename, err)
	}
}

func (sm *SessionManager) LoadFromFile() {
	if sm.filename == "" {
		return
	}

	data, err := os.ReadFile(sm.filename)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("❌ Failed to read sessions from %s: %v", sm.filename, err)
		}
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	var newData sessionData
	if err := json.Unmarshal(data, &newData); err == nil && newData.Sessions != nil {
		sm.sessions = newData.Sessions
		if newData.Active != nil {
			sm.active = newData.Active
		} else {
			sm.active = make(map[int64]string)
		}
		log.Printf("✅ Loaded %d sessions from %s", len(sm.sessions), sm.filename)
		return
	}

	// Fallback: Migrasi format lama (jika sessions.json kamu masih pakai UserID sebagai key)
	var oldSessions map[int64]*Session
	if err := json.Unmarshal(data, &oldSessions); err != nil {
		log.Printf("❌ Failed to parse sessions JSON: %v", err)
		return
	}
	
	for _, s := range oldSessions {
		if s.FullPhone != "" {
			sm.sessions[s.FullPhone] = s
		} else if s.Phone != "" {
			sm.sessions[s.Phone] = s
		}
	}
	sm.active = make(map[int64]string)
	log.Printf("✅ Migrated %d sessions from old format", len(sm.sessions))
}