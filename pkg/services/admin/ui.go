package admin

import "embed"

//go:embed ui/login.html ui/layout.html
var uiFiles embed.FS
