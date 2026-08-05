package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"telkomsel-bot/util"
)

func (m tuiModel) getMenuItems() []string {
	session := m.getActiveSession()
	items := []string{}

	if session == nil {
		items = append(items, "➕ Tambah Akun")
		if len(m.sessions.List()) > 0 {
			items = append(items, "👥 Pilih / Ganti Akun")
		}
	} else {
		items = append(items, "📊 Cek Profil")
		items = append(items, "📦 Cek Kuota")
		items = append(items, "🛒 Beli Paket")
		items = append(items, "⏰ Jadwal Auto-Buy")
		items = append(items, "👥 Ganti Akun")
		items = append(items, "➕ Tambah Akun Baru")
		items = append(items, "👋 Logout")
	}

	return items
}

func (m tuiModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := m.getMenuItems()
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter":
			return m.selectMenu(items[m.cursor])
		}
	}
	return m, nil
}

func (m tuiModel) selectMenu(selected string) (tea.Model, tea.Cmd) {
	m.message = ""
	switch selected {
	case "➕ Tambah Akun", "➕ Tambah Akun Baru":
		m.screen = screenLogin
		m.input.SetValue("")
		m.input.Placeholder = "812xxxxxxxx"
		m.input.Focus()
		return m, nil

	case "👥 Pilih / Ganti Akun", "👥 Ganti Akun":
		m.screen = screenAccountSelect
		m.cursor = 0
		return m, nil

	case "📊 Cek Profil":
		session := m.getActiveSession()
		if session == nil {
			m.message = "Belum memilih akun."
			return m, nil
		}
		m.screen = screenLoading
		m.loading = "Mengambil profil..."
		return m, m.fetchProfile(session)

	case "📦 Cek Kuota":
		session := m.getActiveSession()
		if session == nil {
			m.message = "Belum memilih akun."
			return m, nil
		}
		m.screen = screenLoading
		m.loading = "Mengambil kuota..."
		return m, m.fetchQuota(session)

	case "🛒 Beli Paket":
		session := m.getActiveSession()
		if session == nil {
			m.message = "Belum memilih akun."
			return m, nil
		}
		m.screen = screenLoading
		m.loading = "Mengambil paket..."
		return m, m.fetchOffers(session)

	case "⏰ Jadwal Auto-Buy":
		session := m.getActiveSession()
		if session == nil {
			m.message = "Belum memilih akun."
			return m, nil
		}
		m.screen = screenScheduleMenu
		m.cursor = 0
		return m, nil

	case "👋 Logout":
		session := m.getActiveSession()
		if session == nil {
			m.message = "Belum ada akun aktif."
			return m, nil
		}
		m.sessions.Delete(session.FullPhone)
		m.activeAccount = ""
		m.message = "✓ Sudah logout."
		m.cursor = 0

		accounts := m.sessions.List()
		if len(accounts) > 0 {
			m.screen = screenAccountSelect
		} else {
			m.screen = screenMenu
		}
		return m, nil
	}
	return m, nil
}

func (m tuiModel) updateAccountSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	accounts := m.sessions.List()

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(accounts) { // +1 untuk "Tambah Akun Baru"
				m.cursor++
			}
		case "enter":
			if m.cursor < len(accounts) {
				selectedPhone := accounts[m.cursor].FullPhone
				m.activeAccount = selectedPhone
				m.screen = screenMenu
				m.cursor = 0
				m.message = fmt.Sprintf("✓ Akun +%s dipilih", selectedPhone)
			} else {
				m.screen = screenLogin
				m.input.SetValue("")
				m.input.Placeholder = "812xxxxxxxx"
				m.input.Focus()
			}
			return m, nil
		case "esc":
			if m.getActiveSession() != nil {
				m.screen = screenMenu
				m.cursor = 0
			}
			return m, nil
		}
	}
	return m, nil
}

func (m tuiModel) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "enter" {
			phone := m.input.Value()
			local, full, err := util.ValidatePhone(phone)
			if err != nil {
				m.screen = screenError
				m.message = "Nomor tidak valid."
				return m, nil
			}
			m.loginPhone = local
			m.screen = screenLoading
			m.loading = "Proses login..."
			return m, m.doLogin(local, full)
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) updateOTP(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(autoOtpMsg); ok {
		m.screen = screenLoading
		m.loading = "🤖 Auto OTP diterima, memverifikasi..."
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "enter" {
			otp := m.input.Value()
			if otp != "" && m.otpChan != nil {
				m.otpChan <- otp
				m.screen = screenLoading
				m.loading = "Memverifikasi OTP..."
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) getBuyMenuItems() []string {
	var items []string
	for _, o := range m.offers {
		label := fmt.Sprintf("📦 %s - Rp%s", o.Name, o.Price)
		if len(label) > 60 {
			label = label[:57] + "..."
		}
		items = append(items, label)
	}
	items = append(items, "🆔 Custom Offer ID", "🔙 Kembali")
	return items
}

func (m tuiModel) updateBuyMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := m.getBuyMenuItems()
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter":
			offerCount := len(m.offers)
			switch {
			case m.cursor < offerCount:
				m.buyOfferID = m.offers[m.cursor].ID
				m.screen = screenBuyPayment
				m.cursor = 0
			case m.cursor == offerCount:
				m.screen = screenBuyOfferID
				m.input.SetValue("")
				m.input.Placeholder = "offer ID..."
				m.input.Focus()
			case m.cursor == offerCount+1:
				m.screen = screenMenu
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m tuiModel) updateBuyOfferID(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "enter" {
			m.buyOfferID = m.input.Value()
			if m.buyOfferID == "" {
				m.screen = screenError
				m.message = "Offer ID kosong."
				return m, nil
			}
			m.screen = screenBuyPayment
			m.cursor = 0
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) updateBuyPayment(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := []string{"Pulsa", "QRIS", "Kembali"}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0:
				m.buyPayment = "AIRTIME"
			case 1:
				m.buyPayment = "qris"
			case 2:
				m.screen = screenBuyMenu
				m.cursor = 0
				return m, nil
			}
			session := m.getActiveSession()
			if session == nil {
				m.screen = screenError
				m.message = "Belum memilih akun."
				return m, nil
			}
			m.screen = screenLoading
			m.loading = "Mengambil detail paket..."
			return m, m.fetchPackage(session)
		}
	}
	return m, nil
}

func (m tuiModel) updateBuyConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := []string{"Ya, Beli", "Tidak, Batal"}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == 0 {
				session := m.getActiveSession()
				if session == nil {
					m.screen = screenError
					m.message = "Belum memilih akun."
					return m, nil
				}
				m.screen = screenLoading
				m.loading = "Memproses pembelian..."
				return m, m.doBuy(session)
			}
			m.screen = screenMenu
			m.cursor = 0
		}
	}
	return m, nil
}

func (m tuiModel) updateScheduleMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	session := m.getActiveSession()
	status := "Nonaktif"
	if session != nil && session.AutoBuyActive {
		status = "Aktif"
	}
	items := []string{"Status: " + status, "Ubah Jadwal", "Ubah Threshold", "Ubah Offer ID", "Ubah Payment", "Kembali"}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0:
				if session != nil {
					session.AutoBuyActive = !session.AutoBuyActive
					m.sessions.Set(session.FullPhone, session)
					if session.AutoBuyActive {
						m.message = "✓ Auto-Buy diaktifkan"
					} else {
						m.message = "✓ Auto-Buy dinonaktifkan"
					}
				}
			case 1:
				m.screen = screenScheduleInterval
				m.input.SetValue("")
				m.input.Placeholder = "Contoh: 30"
				m.input.Focus()
				if session != nil && session.AutoBuyInterval > 0 {
					m.input.SetValue(fmt.Sprintf("%d", session.AutoBuyInterval))
				}
				return m, nil
			case 2:
				m.screen = screenScheduleThreshold
				m.input.SetValue("")
				m.input.Placeholder = "Contoh: 500"
				m.input.Focus()
				if session != nil && session.AutoBuyThreshold > 0 {
					m.input.SetValue(fmt.Sprintf("%d", session.AutoBuyThreshold))
				}
				return m, nil
			case 3:
				m.screen = screenScheduleOfferID
				m.input.SetValue("")
				m.input.Placeholder = "Offer ID atau 'ilmupedia'"
				m.input.Focus()
				if session != nil && session.AutoBuyPackage != "" {
					m.input.SetValue(session.AutoBuyPackage)
				}
				return m, nil
			case 4:
				m.screen = screenSchedulePayment
				m.cursor = 0
				return m, nil
			case 5:
				m.screen = screenMenu
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m tuiModel) updateScheduleThreshold(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "enter" {
			schThreshold := m.input.Value()
			session := m.getActiveSession()
			if session != nil {
				var i int
				fmt.Sscanf(schThreshold, "%d", &i)
				session.AutoBuyThreshold = i
				m.sessions.Set(session.FullPhone, session)
				m.message = "✓ Threshold diupdate"
			}
			m.screen = screenScheduleMenu
			m.cursor = 0
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) updateScheduleInterval(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "enter" {
			schInterval := m.input.Value()
			session := m.getActiveSession()
			if session != nil {
				var i int
				fmt.Sscanf(schInterval, "%d", &i)
				session.AutoBuyInterval = i
				m.sessions.Set(session.FullPhone, session)
				m.message = "✓ Interval diupdate"
			}
			m.screen = screenScheduleMenu
			m.cursor = 0
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) updateScheduleOfferID(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "enter" {
			schOfferID := m.input.Value()
			session := m.getActiveSession()
			if session != nil {
				session.AutoBuyPackage = schOfferID
				m.sessions.Set(session.FullPhone, session)
				m.message = "✓ Offer ID diupdate"
			}
			m.screen = screenScheduleMenu
			m.cursor = 0
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) updateSchedulePayment(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := []string{"Pulsa", "QRIS", "Kembali"}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter":
			session := m.getActiveSession()
			switch m.cursor {
			case 0:
				if session != nil {
					session.AutoBuyPayment = "AIRTIME"
					m.sessions.Set(session.FullPhone, session)
					m.message = "✓ Payment diupdate (Pulsa)"
				}
			case 1:
				if session != nil {
					session.AutoBuyPayment = "qris"
					m.sessions.Set(session.FullPhone, session)
					m.message = "✓ Payment diupdate (QRIS)"
				}
			case 2:
			}
			m.screen = screenScheduleMenu
			m.cursor = 0
			return m, nil
		}
	}
	return m, nil
}