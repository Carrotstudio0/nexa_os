package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/governance"
)

const (
	AnalyticsFile = "data/analytics.json"
)

// Session represents a user session
type Session struct {
	ID           string                 `json:"id"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Device       string                 `json:"device"`
	OS           string                 `json:"os"`
	Browser      string                 `json:"browser"`
	ConnectedAt  time.Time              `json:"connected_at"`
	LastActivity time.Time              `json:"last_activity"`
	PageViews    int                    `json:"page_views"`
	Actions      []Action               `json:"actions"`
	Files        []FileActivity         `json:"files"`
	IsActive     bool                   `json:"is_active"`
	Location     string                 `json:"location,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Action represents a user action
type Action struct {
	Type      string                 `json:"type"` // page_view, file_upload, file_download, api_call
	Path      string                 `json:"path"`
	Method    string                 `json:"method,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  int64                  `json:"duration,omitempty"` // milliseconds
	Status    int                    `json:"status,omitempty"`
	Size      int64                  `json:"size,omitempty"` // bytes
	Data      map[string]interface{} `json:"data,omitempty"`
}

// FileActivity represents file operations
type FileActivity struct {
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	Action    string    `json:"action"` // upload, download, delete, view
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // success, failed
}

// AnalyticsManager manages all sessions and analytics
type AnalyticsManager struct {
	sessions   map[string]*Session
	mu         sync.RWMutex
	events     chan Event
	govManager *governance.GovernanceManager
}

// Event represents a real-time event
type Event struct {
	Type      string      `json:"type"`
	SessionID string      `json:"session_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Stats represents overall statistics
type Stats struct {
	TotalSessions          int            `json:"total_sessions"`
	ActiveSessions         int            `json:"active_sessions"`
	TotalPageViews         int            `json:"total_page_views"`
	TotalActions           int            `json:"total_actions"`
	TotalFilesUploaded     int            `json:"total_files_uploaded"`
	TotalFilesDownloaded   int            `json:"total_files_downloaded"`
	AverageSessionDuration float64        `json:"average_session_duration"`
	TopPages               []PageStat     `json:"top_pages"`
	RecentSessions         []*Session     `json:"recent_sessions"`
	DeviceBreakdown        map[string]int `json:"device_breakdown"`
	OSBreakdown            map[string]int `json:"os_breakdown"`
	BrowserBreakdown       map[string]int `json:"browser_breakdown"`
}

// PageStat represents page statistics
type PageStat struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

var globalManager *AnalyticsManager

func init() {
	globalManager = NewAnalyticsManager()
}

// NewAnalyticsManager creates a new analytics manager
func NewAnalyticsManager() *AnalyticsManager {
	am := &AnalyticsManager{
		sessions: make(map[string]*Session),
		events:   make(chan Event, 1000),
	}
	am.Load()
	go am.startPersistence()
	return am
}

func (am *AnalyticsManager) startPersistence() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		am.Save()
	}
}

func (am *AnalyticsManager) Save() {
	am.mu.RLock()
	defer am.mu.RUnlock()

	os.MkdirAll(filepath.Dir(AnalyticsFile), 0755)
	data, err := json.MarshalIndent(am.sessions, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(AnalyticsFile, data, 0644)
}

func (am *AnalyticsManager) Load() {
	am.mu.Lock()
	defer am.mu.Unlock()

	data, err := os.ReadFile(AnalyticsFile)
	if err != nil {
		return
	}

	var loaded map[string]*Session
	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}

	// Mark all sessions as inactive on load
	for _, s := range loaded {
		s.IsActive = false
	}
	am.sessions = loaded
}

func (am *AnalyticsManager) SetGovernance(gm *governance.GovernanceManager) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.govManager = gm
}

// GetManager returns the global analytics manager
func GetManager() *AnalyticsManager {
	return globalManager
}

// CreateSession creates a new session
func (am *AnalyticsManager) CreateSession(id, ip, userAgent string) *Session {
	am.mu.Lock()
	defer am.mu.Unlock()

	device, os, browser := parseUserAgent(userAgent)

	session := &Session{
		ID:           id,
		IPAddress:    ip,
		UserAgent:    userAgent,
		Device:       device,
		OS:           os,
		Browser:      browser,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
		Actions:      []Action{},
		Files:        []FileActivity{},
		Metadata:     make(map[string]interface{}),
	}

	am.sessions[id] = session

	if am.govManager != nil {
		am.govManager.ReportEvent("Network", governance.LevelNotice, "New Session created", fmt.Sprintf("IP: %s, Device: %s", ip, device), "Created")
	}

	// Send event
	am.events <- Event{
		Type:      "session_created",
		SessionID: id,
		Timestamp: time.Now(),
		Data:      session,
	}

	return session
}

// GetSession retrieves a session by ID
func (am *AnalyticsManager) GetSession(id string) *Session {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.sessions[id]
}

// TrackAction tracks a user action
func (am *AnalyticsManager) TrackAction(sessionID string, action Action) {
	am.mu.Lock()
	defer am.mu.Unlock()

	session, exists := am.sessions[sessionID]
	if !exists {
		return
	}

	action.Timestamp = time.Now()
	session.Actions = append(session.Actions, action)
	session.LastActivity = time.Now()

	if action.Type == "page_view" {
		session.PageViews++
	}

	// Security Check: Too many actions in short time
	if len(session.Actions) > 100 && am.govManager != nil {
		recentCount := 0
		for i := len(session.Actions) - 1; i >= 0 && i > len(session.Actions)-50; i-- {
			if time.Since(session.Actions[i].Timestamp) < 10*time.Second {
				recentCount++
			}
		}
		if recentCount > 30 {
			am.govManager.ReportEvent("Security", governance.LevelAction, "Session Abnormality",
				fmt.Sprintf("Session %s generated %d actions in 10s", sessionID, recentCount), "Throttling")
		}
	}

	// Send event
	am.events <- Event{
		Type:      "action",
		SessionID: sessionID,
		Timestamp: time.Now(),
		Data:      action,
	}
}

// TrackFile tracks file activity
func (am *AnalyticsManager) TrackFile(sessionID string, file FileActivity) {
	am.mu.Lock()
	defer am.mu.Unlock()

	session, exists := am.sessions[sessionID]
	if !exists {
		return
	}

	file.Timestamp = time.Now()
	session.Files = append(session.Files, file)
	session.LastActivity = time.Now()

	// Send event
	am.events <- Event{
		Type:      "file_activity",
		SessionID: sessionID,
		Timestamp: time.Now(),
		Data:      file,
	}
}

// GetActiveSessions returns all active sessions
func (am *AnalyticsManager) GetActiveSessions() []*Session {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var active []*Session
	for _, session := range am.sessions {
		if session.IsActive && time.Since(session.LastActivity) < 5*time.Minute {
			active = append(active, session)
		}
	}
	return active
}

// GetAllSessions returns all sessions
func (am *AnalyticsManager) GetAllSessions() []*Session {
	am.mu.RLock()
	defer am.mu.RUnlock()

	sessions := make([]*Session, 0, len(am.sessions))
	for _, s := range am.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// GetStats returns overall statistics
func (am *AnalyticsManager) GetStats() Stats {
	am.mu.RLock()
	defer am.mu.RUnlock()

	stats := Stats{
		TotalSessions:    len(am.sessions),
		TopPages:         []PageStat{},
		RecentSessions:   []*Session{},
		DeviceBreakdown:  make(map[string]int),
		OSBreakdown:      make(map[string]int),
		BrowserBreakdown: make(map[string]int),
	}

	pageViews := make(map[string]int)
	var totalDuration float64
	var filesUploaded, filesDownloaded int

	for _, session := range am.sessions {
		if session.IsActive && time.Since(session.LastActivity) < 5*time.Minute {
			stats.ActiveSessions++
		}
		stats.TotalPageViews += session.PageViews
		stats.TotalActions += len(session.Actions)

		// Calculate duration
		duration := time.Since(session.ConnectedAt).Seconds()
		totalDuration += duration

		// Count pages
		for _, action := range session.Actions {
			if action.Type == "page_view" {
				pageViews[action.Path]++
			}
		}

		// Count files
		for _, file := range session.Files {
			switch file.Action {
			case "upload":
				filesUploaded++
			case "download":
				filesDownloaded++
			}
		}

		// Device/OS/Browser breakdown
		stats.DeviceBreakdown[session.Device]++
		stats.OSBreakdown[session.OS]++
		stats.BrowserBreakdown[session.Browser]++
	}

	stats.TotalFilesUploaded = filesUploaded
	stats.TotalFilesDownloaded = filesDownloaded

	if len(am.sessions) > 0 {
		stats.AverageSessionDuration = totalDuration / float64(len(am.sessions))
	}

	// Top pages
	for path, count := range pageViews {
		stats.TopPages = append(stats.TopPages, PageStat{Path: path, Count: count})
	}

	// Get recent sessions (last 10)
	count := 0
	for _, session := range am.sessions {
		if count >= 10 {
			break
		}
		stats.RecentSessions = append(stats.RecentSessions, session)
		count++
	}

	return stats
}

// CloseSession marks a session as inactive
func (am *AnalyticsManager) CloseSession(sessionID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if session, exists := am.sessions[sessionID]; exists {
		session.IsActive = false

		am.events <- Event{
			Type:      "session_closed",
			SessionID: sessionID,
			Timestamp: time.Now(),
			Data:      session,
		}
	}
}

// GetEventChannel returns the events channel
func (am *AnalyticsManager) GetEventChannel() <-chan Event {
	return am.events
}

// ToJSON converts stats to JSON
func (s Stats) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// Simple user agent parser
func parseUserAgent(ua string) (device, os, browser string) {
	// Very basic parsing - можно улучшить с помощью библиотеки
	device = "Desktop"
	if contains(ua, "Mobile") || contains(ua, "Android") || contains(ua, "iPhone") {
		device = "Mobile"
	} else if contains(ua, "Tablet") || contains(ua, "iPad") {
		device = "Tablet"
	}

	os = "Unknown"
	if contains(ua, "Windows") {
		os = "Windows"
	} else if contains(ua, "Mac") {
		os = "macOS"
	} else if contains(ua, "Linux") {
		os = "Linux"
	} else if contains(ua, "Android") {
		os = "Android"
	} else if contains(ua, "iOS") || contains(ua, "iPhone") {
		os = "iOS"
	}

	browser = "Unknown"
	if contains(ua, "Chrome") && !contains(ua, "Edge") {
		browser = "Chrome"
	} else if contains(ua, "Firefox") {
		browser = "Firefox"
	} else if contains(ua, "Safari") && !contains(ua, "Chrome") {
		browser = "Safari"
	} else if contains(ua, "Edge") {
		browser = "Edge"
	}

	return
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
