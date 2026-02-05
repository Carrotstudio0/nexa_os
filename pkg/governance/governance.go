package governance

import (
	"fmt"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/utils"
)

// DecisionLevel represents the severity of a governance decision
type DecisionLevel string

const (
	LevelNotice   DecisionLevel = "notice"
	LevelWarning  DecisionLevel = "warning"
	LevelCritical DecisionLevel = "critical"
	LevelAction   DecisionLevel = "action"
)

// GovernanceEvent records what happened and why
type GovernanceEvent struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Type      string        `json:"type"`
	Severity  DecisionLevel `json:"severity"`
	Message   string        `json:"message"`
	Reason    string        `json:"reason"`
	Action    string        `json:"action_taken,omitempty"`
}

// GovernanceManager is the "Brain" of Nexa
type GovernanceManager struct {
	mu           sync.RWMutex
	PolicyEngine *PolicyEngine
	Timeline     []GovernanceEvent
	networkMgr   *network.NetworkManager

	onAction func(event GovernanceEvent)
}

func NewGovernanceManager(pe *PolicyEngine, nm *network.NetworkManager) *GovernanceManager {
	return &GovernanceManager{
		PolicyEngine: pe,
		networkMgr:   nm,
		Timeline:     make([]GovernanceEvent, 0),
	}
}

// AnalyzeSystem state and make decisions
func (gm *GovernanceManager) AnalyzeSystem() {
	if gm.networkMgr == nil {
		return
	}

	policy := gm.PolicyEngine.GetPolicy()
	topo := gm.networkMgr.GetTopology()

	// 1. Check Client Load
	if len(topo.Devices) > policy.MaxClients {
		gm.Decide(GovernanceEvent{
			Type:     "Network",
			Severity: LevelWarning,
			Message:  "Network capacity reaching limits",
			Reason:   fmt.Sprintf("Active devices (%d) exceeds policy max (%d)", len(topo.Devices), policy.MaxClients),
			Action:   "Restricting new connections",
		})
	}

	// 2. Check Latency Spikes
	for _, device := range topo.Devices {
		if device.IsOnline && device.Metrics.LatencyMS > policy.LatencyThreshold {
			gm.Decide(GovernanceEvent{
				Type:     "Performance",
				Severity: LevelAction,
				Message:  fmt.Sprintf("High latency detected on node: %s", device.Name),
				Reason:   fmt.Sprintf("Latency %dms exceeds threshold %dms", device.Metrics.LatencyMS, policy.LatencyThreshold),
				Action:   "Throttle/Optimize Route",
			})
		}

		if device.IsOnline && device.Metrics.ErrorRate > policy.ErrorRateLimit {
			gm.Decide(GovernanceEvent{
				Type:     "Stability",
				Severity: LevelCritical,
				Message:  fmt.Sprintf("Critical error rate on node: %s", device.Name),
				Reason:   fmt.Sprintf("Error rate %.2f%% exceeds limit %.2f%%", device.Metrics.ErrorRate, policy.ErrorRateLimit),
				Action:   "Quarantine Node",
			})
		}
	}
}

func (gm *GovernanceManager) Decide(event GovernanceEvent) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	event.Timestamp = time.Now()

	// Avoid flooding timeline with same message frequently
	if len(gm.Timeline) > 0 {
		last := gm.Timeline[len(gm.Timeline)-1]
		if last.Message == event.Message && time.Since(last.Timestamp) < 1*time.Minute {
			return
		}
	}

	gm.Timeline = append(gm.Timeline, event)
	if len(gm.Timeline) > 100 {
		gm.Timeline = gm.Timeline[1:] // Keep last 100 events
	}

	utils.LogInfo("Governance", fmt.Sprintf("[%s] %s: %s", event.Severity, event.Type, event.Message))

	if gm.onAction != nil {
		gm.onAction(event)
	}
}

func (gm *GovernanceManager) SetOnAction(callback func(GovernanceEvent)) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.onAction = callback
}

func (gm *GovernanceManager) GetTimeline() []GovernanceEvent {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	// Return a copy
	events := make([]GovernanceEvent, len(gm.Timeline))
	copy(events, gm.Timeline)
	return events
}

// ReportEvent allows services to report incidents to the governance engine
func (gm *GovernanceManager) ReportEvent(eType string, severity DecisionLevel, message, reason, action string) {
	gm.Decide(GovernanceEvent{
		Type:     eType,
		Severity: severity,
		Message:  message,
		Reason:   reason,
		Action:   action,
	})
}

func (gm *GovernanceManager) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			gm.AnalyzeSystem()
		}
	}()
}
