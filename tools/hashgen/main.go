package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Simple utility to generate bcrypt hashes for passwords
// Usage: go run tools/hashgen.go

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("========================================")
	fmt.Println("Nexa Protocol - Password Hash Generator")
	fmt.Println("========================================")
	fmt.Println()

	for {
		fmt.Print("Enter password (or 'quit' to exit): ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		if password == "quit" || password == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if password == "" {
			fmt.Println("⚠️  Password cannot be empty")
			continue
		}

		// Generate bcrypt hash
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Println()
		fmt.Println("✅ Hash generated successfully!")
		fmt.Println("---")
		fmt.Printf("Password: %s\n", password)
		fmt.Printf("Hash:     %s\n", string(hash))
		fmt.Println("---")
		fmt.Println()
		fmt.Println("Add this to users.json:")
		fmt.Println(`{`)
		fmt.Printf(`  "username": {` + "\n")
		fmt.Printf(`    "password": "%s",`+"\n", string(hash))
		fmt.Printf(`    "role": "user"` + "\n")
		fmt.Printf(`  }` + "\n")
		fmt.Println(`}`)
		fmt.Println()
	}
}
