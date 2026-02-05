package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Block represents a single data record in the chain
type Block struct {
	Index        int    `json:"index"`
	Timestamp    string `json:"timestamp"`
	Key          string `json:"key"`
	Value        string `json:"value"`
	PreviousHash string `json:"previous_hash"`
	Hash         string `json:"hash"`
	Validator    string `json:"validator"` // Node that validated this
}

// Blockchain manages the chain of blocks
type Blockchain struct {
	mu       sync.RWMutex
	Chain    []Block           `json:"chain"`
	Data     map[string]string `json:"data"` // Quick lookup
	Filename string            `json:"-"`
}

// NewBlockchain initializes the ledger
func NewBlockchain(filename string) (*Blockchain, error) {
	bc := &Blockchain{
		Chain:    []Block{},
		Data:     make(map[string]string),
		Filename: filename,
	}

	// Load from disk
	if _, err := os.Stat(filename); err == nil {
		content, err := os.ReadFile(filename)
		if err == nil {
			json.Unmarshal(content, bc)
			// Rebuild Data map
			for _, b := range bc.Chain {
				bc.Data[b.Key] = b.Value
			}
			return bc, nil
		}
	}

	// Genesis Block
	bc.AddBlock("genesis", "Nexa Protocol Genesis Block", "SYSTEM")
	return bc, nil
}

// CalculateHash generates a SHA256 hash for a block
func CalculateHash(b Block) string {
	record := fmt.Sprintf("%d%s%s%s%s%s", b.Index, b.Timestamp, b.Key, b.Value, b.PreviousHash, b.Validator)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// AddBlock adds a new data block to the chain
func (bc *Blockchain) AddBlock(key, value, validator string) Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	var prevHash string
	if len(bc.Chain) > 0 {
		prevHash = bc.Chain[len(bc.Chain)-1].Hash
	}

	newBlock := Block{
		Index:        len(bc.Chain),
		Timestamp:    time.Now().Format(time.RFC3339),
		Key:          key,
		Value:        value,
		PreviousHash: prevHash,
		Validator:    validator,
	}
	newBlock.Hash = CalculateHash(newBlock)

	bc.Chain = append(bc.Chain, newBlock)
	bc.Data[key] = value // Update quick lookup
	bc.save()

	return newBlock
}

// Get retrieves a value (latest state)
func (bc *Blockchain) Get(key string) (string, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	val, ok := bc.Data[key]
	return val, ok
}

// IsChainValid checks integrity
func (bc *Blockchain) IsChainValid() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.Chain); i++ {
		current := bc.Chain[i]
		prev := bc.Chain[i-1]

		if current.Hash != CalculateHash(current) {
			return false
		}
		if current.PreviousHash != prev.Hash {
			return false
		}
	}
	return true
}

func (bc *Blockchain) save() {
	data, _ := json.MarshalIndent(bc, "", "  ")
	os.WriteFile(bc.Filename, data, 0644)
}
