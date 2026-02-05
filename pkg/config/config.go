package config

// Global System Configuration
const (
	GatewayPort   = "8000"
	AdminPort     = "8080"
	WebPort       = "8081"
	StoragePort   = "8081" // Alias for WebPort
	ChatPort      = "8082"
	DashboardPort = "7000"
	ServerPort    = "1413"
	DNSPort       = "1112"

	GatewayTarget   = "http://127.0.0.1:8000"
	AdminTarget     = "http://127.0.0.1:8080"
	WebTarget       = "http://127.0.0.1:8081"
	DashboardTarget = "http://127.0.0.1:7000"
	ChatTarget      = "http://127.0.0.1:8082"
)

// Service Metadata
var Services = []map[string]string{
	{"name": "Admin Panel", "url": "/admin", "port": AdminPort, "desc": "إدارة وتحسين النظام والتحكم في المستخدمين", "icon": "settings"},
	{"name": "File Manager", "url": "/storage", "port": StoragePort, "desc": "إدارة الملفات والأرشيف الرقمي المشترك", "icon": "folder"},
	{"name": "Quantum Chat", "url": "/chat", "port": ChatPort, "desc": "نظام التواصل الفوري المشفر بين الأجهزة", "icon": "chat"},
	{"name": "Dashboard", "url": "/dashboard", "port": DashboardPort, "desc": "لوحة المراقبة الرئيسية لإحصائيات النظام", "icon": "dashboard"},
}
