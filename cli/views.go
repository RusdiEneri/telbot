package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdp/qrterminal/v3"

	"telkomsel-bot/util"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4")).
			Bold(true)

	menuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F7FFF7"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2ECC71")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E74C3C")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3498DB")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3498DB")).
			Padding(1, 2)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95A5A6"))
)

func (m tuiModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("📱 Telbot v%s", util.GetVersion())))
	b.WriteString("\n\n")

	switch m.screen {
	case screenMenu:
		b.WriteString(m.viewMenu())
	case screenAccountSelect:
		b.WriteString(m.viewAccountSelect())
	case screenLogin:
		b.WriteString(m.viewInput("Masukkan Nomor HP", "Contoh: 812xxxxxxxx"))
	case screenOTP:
		b.WriteString(m.viewInput("Masukkan Kode OTP", "OTP dikirim via SMS"))
	case screenLoading:
		b.WriteString(m.viewLoading())
	case screenProfile:
		b.WriteString(m.viewResult("Profil"))
	case screenQuota:
		b.WriteString(m.viewResult("Kuota"))
	case screenBuyMenu:
		b.WriteString(m.viewBuyMenu())
	case screenBuyOfferID:
		b.WriteString(m.viewInput("Masukkan Offer ID", "ID paket dari my.telkomsel.com"))
	case screenBuyPayment:
		b.WriteString(m.viewBuyPayment())
	case screenBuyConfirm:
		b.WriteString(m.viewBuyConfirm())
	case screenBuyResult:
		b.WriteString(m.viewResult("Hasil Pembelian"))
	case screenPaymentPoll:
		b.WriteString(m.viewPaymentPoll())
	case screenScheduleMenu:
		b.WriteString(m.viewScheduleMenu())
	case screenScheduleInterval:
		b.WriteString(m.viewInput("Masukkan Interval (menit)", "Contoh: 15, 30, 60"))
	case screenScheduleThreshold:
		b.WriteString(m.viewInput("Masukkan Threshold (MB)", "Contoh: 0 (Habis), 500"))
	case screenScheduleOfferID:
		b.WriteString(m.viewInput("Masukkan Offer ID", "ID paket atau 'ilmupedia'"))
	case screenSchedulePayment:
		b.WriteString(m.viewSchedulePayment())
	case screenError:
		b.WriteString(m.viewError())
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("esc: kembali • q: keluar"))
	b.WriteString("\n")

	return b.String()
}

func (m tuiModel) viewMenu() string {
	var b strings.Builder
	items := m.getMenuItems()

	session := m.getActiveSession()
	if session != nil {
		b.WriteString(successStyle.Render("● Logged in: +" + session.FullPhone))
		b.WriteString("\n\n")
	} else {
		b.WriteString(dimStyle.Render("● Belum memilih akun"))
		b.WriteString("\n\n")
	}

	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + item))
		} else {
			b.WriteString(menuStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}

	if m.message != "" {
		b.WriteString("\n")
		if strings.HasPrefix(m.message, "✓") {
			b.WriteString(successStyle.Render(m.message))
		} else {
			b.WriteString(errorStyle.Render("✗ " + m.message))
		}
	}

	return b.String()
}

func (m tuiModel) viewInput(title, hint string) string {
	var b strings.Builder
	b.WriteString(infoStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(hint))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Tekan Enter untuk melanjutkan"))
	return b.String()
}

func (m tuiModel) viewLoading() string {
	return fmt.Sprintf("%s %s", m.spinner.View(), infoStyle.Render(m.loading))
}

func (m tuiModel) viewResult(title string) string {
	var b strings.Builder
	b.WriteString(infoStyle.Render("━━ " + title + " ━━"))
	b.WriteString("\n\n")
	b.WriteString(m.result)
	return b.String()
}

func (m tuiModel) viewError() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("❌ Error"))
	b.WriteString("\n\n")
	b.WriteString(m.message)
	return b.String()
}

func (m tuiModel) viewBuyMenu() string {
	items := m.getBuyMenuItems()
	var b strings.Builder
	b.WriteString(infoStyle.Render(fmt.Sprintf("📦 Pilih Paket (%d rekomendasi)", len(m.offers))))
	b.WriteString("\n\n")
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + item))
		} else {
			b.WriteString(menuStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m tuiModel) viewBuyPayment() string {
	items := []string{"💰 Pulsa", "📱 QRIS", "🔙 Kembali"}
	var b strings.Builder
	b.WriteString(infoStyle.Render("Metode Pembayaran"))
	b.WriteString("\n\n")
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + item))
		} else {
			b.WriteString(menuStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m tuiModel) viewBuyConfirm() string {
	var b strings.Builder
	if m.buyDetails != nil {
		b.WriteString(infoStyle.Render("━━ Detail Paket ━━"))
		b.WriteString("\n\n")
		detail := fmt.Sprintf("Nama:  %s\nHarga: Rp%s\nMasa:  %s", m.buyDetails.Name, m.buyDetails.Price, m.buyDetails.Validity)
		b.WriteString(boxStyle.Render(detail))
		b.WriteString("\n\n")
	}

	b.WriteString(infoStyle.Render("Lanjutkan pembelian?"))
	b.WriteString("\n\n")

	items := []string{"✅ Ya, Beli", "❌ Tidak, Batal"}
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + item))
		} else {
			b.WriteString(menuStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m tuiModel) viewPaymentPoll() string {
	var b strings.Builder
	b.WriteString(infoStyle.Render("📱 Pembayaran QRIS"))
	b.WriteString("\n\n")
	b.WriteString(m.result)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s Menunggu pembayaran (status: %s)...", m.spinner.View(), m.pollStatus))
	return b.String()
}

func (m tuiModel) viewScheduleMenu() string {
	b := strings.Builder{}
	b.WriteString(infoStyle.Render("⏱️ Setup Schedule Auto-Buy"))
	b.WriteString("\n\n")

	session := m.getActiveSession()
	if session != nil {
		interval := fmt.Sprintf("%d menit", session.AutoBuyInterval)
		if session.AutoBuyInterval == 0 {
			interval = "(belum diset)"
		}
		thresholdStr := fmt.Sprintf("< %d MB", session.AutoBuyThreshold)
		if session.AutoBuyThreshold == 0 {
			thresholdStr = "0 MB (Habis)"
		}
		offer := session.AutoBuyPackage
		if offer == "" {
			offer = "(belum diset)"
		}
		pay := session.AutoBuyPayment
		if pay == "" {
			pay = "(belum diset)"
		}
		s := fmt.Sprintf("Interval : %s\nThreshold: %s\nOffer ID : %s\nPayment  : %s", interval, thresholdStr, offer, pay)
		b.WriteString(boxStyle.Render(s))
		b.WriteString("\n\n")
	}

	status := "Nonaktif"
	if session != nil && session.AutoBuyActive {
		status = "Aktif"
	}
	items := []string{"Status: " + status, "Ubah Jadwal", "Ubah Threshold", "Ubah Offer ID", "Ubah Payment", "🔙 Kembali"}
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + item))
		} else {
			b.WriteString(menuStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}

	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render(m.message))
	}
	return b.String()
}

func (m tuiModel) viewAccountSelect() string {
	var b strings.Builder

	accounts := m.sessions.List()

	b.WriteString(infoStyle.Render("━━ Pilih / Ganti Akun ━━"))
	b.WriteString("\n\n")

	activeSession := m.getActiveSession()
	for i, acc := range accounts {
		phone := acc.FullPhone
		if phone == "" {
			phone = acc.Phone
		}

		label := fmt.Sprintf("📱 +%s", phone)
		if activeSession != nil && activeSession.FullPhone == acc.FullPhone {
			label += " (Aktif)"
		} else if acc.IsLoggedIn() {
			label += " ✓"
		} else {
			label += " (expired)"
		}

		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + label))
		} else {
			b.WriteString(menuStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	label := "➕ Tambah Akun Baru"
	if m.cursor == len(accounts) {
		b.WriteString(selectedStyle.Render("▸ " + label))
	} else {
		b.WriteString(menuStyle.Render("  " + label))
	}
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("↑/↓: navigasi • enter: pilih • esc: kembali"))

	return b.String()
}

func (m tuiModel) viewSchedulePayment() string {
	items := []string{"💰 Pulsa", "📱 QRIS", "🔙 Kembali"}
	var b strings.Builder
	b.WriteString(infoStyle.Render("Metode Pembayaran Default"))
	b.WriteString("\n\n")
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("▸ " + item))
		} else {
			b.WriteString(menuStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func renderQR(url string) string {
	var b strings.Builder
	config := qrterminal.Config{
		Level:      qrterminal.L,
		Writer:     &b,
		HalfBlocks: true,
		BlackChar:  "█",
		WhiteChar:  " ",
		QuietZone:  1,
	}
	qrterminal.GenerateWithConfig(url, config)
	return b.String()
}