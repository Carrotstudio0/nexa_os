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
	ColorDim    = "\033[2m"
)

func LogInfo(module, message string) {
	fmt.Printf("%s%s %s%-12s%s %s%s%s %s\n",
		ColorDim, time.Now().Format("15:04:05.000"),
		ColorCyan+ColorBold, "["+module+"]", ColorReset,
		ColorBlue, "➜", ColorReset,
		message)
}

func LogSuccess(module, message string) {
	fmt.Printf("%s%s %s%-12s%s %s%s%s %s%s%s\n",
		ColorDim, time.Now().Format("15:04:05.000"),
		ColorGreen+ColorBold, "["+module+"]", ColorReset,
		ColorGreen, "✔", ColorReset,
		ColorGreen+ColorBold, message, ColorReset)
}

func LogWarning(module, message string) {
	fmt.Printf("%s%s %s%-12s%s %s%s%s %s%s%s\n",
		ColorDim, time.Now().Format("15:04:05.000"),
		ColorYellow+ColorBold, "["+module+"]", ColorReset,
		ColorYellow, "⚡", ColorReset,
		ColorYellow, message, ColorReset)
}

func LogError(module, message string, err error) {
	fmt.Printf("%s%s %s%-12s%s %s%s%s %s%s%s: %v\n",
		ColorDim, time.Now().Format("15:04:05.000"),
		ColorRed+ColorBold, "["+module+"]", ColorReset,
		ColorRed, "✖", ColorReset,
		ColorRed+ColorBold, message, ColorReset, err)
}

func LogFatal(module, message string) {
	fmt.Printf("%s%s %s%-12s%s %s%s%s %s%s%s\n",
		ColorDim, time.Now().Format("15:04:05.000"),
		ColorRed+ColorBold, "["+module+"]", ColorReset,
		ColorRed, "💀", ColorReset,
		ColorRed+ColorBold, message, ColorReset)
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
	banner := `
   _   _  _______   __   _      _   _ _    _____ ___ __  __   _  _____ _____ 
  | \ | || ____\ \ / /  / \    | | | | |  |_   _|_ _|  \/  | / \|_   _| ____|
  |  \| ||  _|  \ V /  / _ \   | | | | |    | |  | || |\/| |/ _ \ | | |  _|  
  | |\  || |___ /   \ / ___ \  | |_| | |___ | |  | || |  | / ___ \| | | |___ 
  |_| \_||_____/_/ \_/_/   \_\  \___/|_____||_| |___|_|  |_/_/   \_\_| |_____|
                                                                              
`
	fmt.Print(ColorCyan + ColorBold + banner + ColorReset)
	fmt.Printf("   %s%s %s%s %s| Integration: %sEnabled%s | Status: %sOnline%s\n",
		ColorWhite+ColorBold, name, ColorCyan, version, ColorReset,
		ColorGreen, ColorReset, ColorGreen, ColorReset)
	fmt.Println(ColorDim + "   ══════════════════════════════════════════════════════════════════════════════" + ColorReset)
}
