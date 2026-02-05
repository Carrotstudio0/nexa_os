package utils

import (
	"fmt"
	"os"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

func LogInfo(module, message string) {
	fmt.Printf("%s[%s]%s %s%-10s%s %s\n",
		ColorBlue, time.Now().Format("15:04:05"), ColorReset,
		ColorCyan+ColorBold, "["+module+"]", ColorReset,
		message)
}

func LogSuccess(module, message string) {
	fmt.Printf("%s[%s]%s %s%-10s%s %s✅ %s%s\n",
		ColorBlue, time.Now().Format("15:04:05"), ColorReset,
		ColorGreen+ColorBold, "["+module+"]", ColorReset,
		ColorGreen, message, ColorReset)
}

func LogWarning(module, message string) {
	fmt.Printf("%s[%s]%s %s%-10s%s %s⚠️  %s%s\n",
		ColorBlue, time.Now().Format("15:04:05"), ColorReset,
		ColorYellow+ColorBold, "["+module+"]", ColorReset,
		ColorYellow, message, ColorReset)
}

func LogError(module, message string, err error) {
	fmt.Printf("%s[%s]%s %s%-10s%s %s❌ %s: %v%s\n",
		ColorBlue, time.Now().Format("15:04:05"), ColorReset,
		ColorRed+ColorBold, "["+module+"]", ColorReset,
		ColorRed, message, err, ColorReset)
}

func LogFatal(module, message string) {
	fmt.Printf("%s[%s]%s %s%-10s%s %s💀 %s%s\n",
		ColorBlue, time.Now().Format("15:04:05"), ColorReset,
		ColorRed+ColorBold, "["+module+"]", ColorReset,
		ColorRed, message, ColorReset)
	os.Exit(1)
}

func SaveEndpoint(name string, url string) {
	dir := "data/endpoints"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	filePath := fmt.Sprintf("%s/%s.url", dir, name)
	os.WriteFile(filePath, []byte(url), 0644)
}

func PrintBanner(name, version string) {
	fmt.Printf("%s%s %s Ultimate%s\n", ColorCyan+ColorBold, name, version, ColorReset)
	fmt.Println("   ════════════════════════════════════════════════════════════════")
}
