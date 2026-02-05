package auth

import (
	"encoding/json"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

type AuthManager struct {
	mu       sync.RWMutex
	Users    map[string]*User `json:"users"`
	filename string
}

func NewAuthManager(filename string) (*AuthManager, error) {
	am := &AuthManager{
		Users:    make(map[string]*User),
		filename: filename,
	}

	if _, err := os.Stat(filename); err == nil {
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(content, &am.Users); err != nil {
			return nil, err
		}
	} else {
		// Create default admin if file doesn't exist
		pass, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		am.Users["admin"] = &User{
			Password: string(pass),
			Role:     "admin",
		}
		am.save()
	}

	return am, nil
}

func (am *AuthManager) Verify(username, password string) (bool, string) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.Users[username]
	if !exists {
		return false, ""
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false, ""
	}

	return true, user.Role
}

func (am *AuthManager) save() error {
	data, err := json.MarshalIndent(am.Users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(am.filename, data, 0644)
}
