package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"

	"telkomsel-bot/config"
	"telkomsel-bot/model"
	"telkomsel-bot/otp"
	"telkomsel-bot/telkomsel"
)

func Run() {
	logPath := filepath.Join(os.TempDir(), "telkomsel-cli.log")
	logFile, err := os.Create(logPath)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	} else {
		log.SetOutput(io.Discard)
	}

	api := telkomsel.NewClient()
	defer api.Close()
	auth := telkomsel.NewAuth()
	sessions := model.NewSessionManager(config.GetSessionPath())

	m := newModel(api, auth, sessions)

	// Start OTP webhook listener if configured
	otpPort := 0
	if portStr := os.Getenv("OTP_WEBHOOK_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			otpPort = p
		}
	}
	if otpPort > 0 {
		otpSecret := os.Getenv("OTP_WEBHOOK_SECRET")
		listener := otp.NewListener(otpPort, otpSecret)
		if err := listener.Start(); err != nil {
			log.Printf("Failed to start OTP listener: %v", err)
		} else {
			m.otpListener = listener
			log.Printf("OTP listener on :%d", otpPort)
		}
	}

	// Auto-select last used account or show account selector
	accounts := sessions.List()
	if len(accounts) > 0 {
		var latestSession *model.Session
		var latestTime int64
		for _, s := range accounts {
			if s.IsLoggedIn() && s.LastLoginAt.Unix() > latestTime {
				latestTime = s.LastLoginAt.Unix()
				latestSession = s
			}
		}

		if latestSession != nil {
			m.activeAccount = latestSession.FullPhone
			m.screen = screenMenu
		} else {
			m.screen = screenAccountSelect
		}
	} else {
		m.screen = screenMenu
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	programRef = p
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}