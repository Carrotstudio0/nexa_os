package governance

import (
	"encoding/json"
	"os"
	"sync"
)

// Policy defines the operational limits and rules of the Nexa system
type Policy struct {
	MaxClients        int     `json:"network.max_clients"`
	MaxUploadSizeMB   int     `json:"storage.max_upload_mb"`
	ChatRateLimit     int     `json:"chat.rate_limit"`       // msgs per sec
	LatencyThreshold  int64   `json:"network.latency_limit"` // ms
	ErrorRateLimit    float64 `json:"network.error_limit"`   // percentage
	AutoRestartFailed bool    `json:"system.auto_restart"`
	QuietHoursEnabled bool    `json:"system.quiet_hours"`
}

// PolicyEngine manages the system's constitution
type PolicyEngine struct {
	mu         sync.RWMutex
	current    Policy
	policyPath string
}

func NewPolicyEngine(path string) *PolicyEngine {
	pe := &PolicyEngine{
		policyPath: path,
		current: Policy{
			MaxClients:        50,
			MaxUploadSizeMB:   1024,
			ChatRateLimit:     5,
			LatencyThreshold:  500,
			ErrorRateLimit:    5.0,
			AutoRestartFailed: true,
		},
	}
	pe.Load()
	return pe
}

func (pe *PolicyEngine) Load() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	data, err := os.ReadFile(pe.policyPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &pe.current)
}

func (pe *PolicyEngine) Save() error {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	data, err := json.MarshalIndent(pe.current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pe.policyPath, data, 0644)
}

func (pe *PolicyEngine) GetPolicy() Policy {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.current
}

func (pe *PolicyEngine) UpdatePolicy(p Policy) {
	pe.mu.Lock()
	pe.current = p
	pe.mu.Unlock()
	pe.Save()
}
