package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	appDir := filepath.Join(configDir, "telbot")
	os.MkdirAll(appDir, 0755)
	return appDir
}

func GetSessionPath() string {
	return filepath.Join(GetConfigDir(), "sessions.json")
}

const (
	BaseURL            = "https://tdw.telkomsel.com"
	QuotaEndpoint      = "/api/subscriber/v5/bonuses"
	BuyEndpoint        = "/api/payment/fulfillment/v2"
	StatusEndpoint     = "/api/payment/status"
	OffersEndpoint     = "/api/offers/recommended/v2"
	LoginURL           = "https://my.telkomsel.com"
	OfferId            = "0fc00fd41bcd26376d806925d746705e"
	DefaultPayment     = "qris"
	MaxRetries         = 3
	WebAppVersion      = "2.0.0"
	ChromePreset       = "chrome-145"
	CiamBaseURL        = "https://ciam.telkomsel.com"
	CiamRealm          = "tsel"
	ClientID           = "e7126474617aa39eb9e484233c9b0649"
	ClientSecret       = "P@ssw0rd"
	RedirectURI        = "https://my.telkomsel.com/web/callback"
	AuthUserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	EncryptionPassword = "production"
)

var Verbose bool

type Config struct {
	BotToken         string
	AdminID          int64
	OTPWebhookPort   int
	OTPWebhookSecret string
}

func Load() *Config {
	// 1. Coba load .env dari Current Working Directory (Folder root 'telbot')
	// Ini akan berjalan jika kamu run dari dalam folder telbot (misal: go run main.go)
	if err := godotenv.Load(".env"); err != nil {
		if Verbose {
			log.Printf("ℹ️  No .env file found in current directory: %v", err)
		}
	}

	// 2. Fallback: Coba load .env dari lokasi file executable (binary)
	// Berguna kalau kamu sudah compile (go build) dan run binary-nya dari mana saja
	if ex, err := os.Executable(); err == nil {
		exDir := filepath.Dir(ex)
		envExPath := filepath.Join(exDir, ".env")
		godotenv.Load(envExPath) // Load jika ada, abaikan jika error
	}

	// 3. Fallback: Load .env dari User Config Dir (~/.config/telbot/.env)
	envPath := filepath.Join(GetConfigDir(), ".env")
	if err := godotenv.Load(envPath); err != nil {
		if Verbose {
			log.Printf("ℹ️  No .env file found at %s: %v", envPath, err)
		}
	}

	// Validasi Env Vars
	token := os.Getenv("TELKOMSEL_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ TELKOMSEL_BOT_TOKEN env var is required")
	}

	adminIDStr := os.Getenv("TELEGRAM_ADMIN_ID")
	var adminID int64
	if adminIDStr != "" {
		if id, err := strconv.ParseInt(adminIDStr, 10, 64); err == nil {
			adminID = id
		} else {
			log.Fatalf("❌ Invalid TELEGRAM_ADMIN_ID: %v", err)
		}
	} else {
		log.Fatal("❌ TELEGRAM_ADMIN_ID env var is required")
	}

	otpPort := 0
	if portStr := os.Getenv("OTP_WEBHOOK_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			otpPort = p
		}
	}
	otpSecret := os.Getenv("OTP_WEBHOOK_SECRET")

	return &Config{
		BotToken:         token,
		AdminID:          adminID,
		OTPWebhookPort:   otpPort,
		OTPWebhookSecret: otpSecret,
	}
}