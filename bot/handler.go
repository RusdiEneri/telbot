package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"telkomsel-bot/model"
	"telkomsel-bot/otp"
	"telkomsel-bot/telkomsel"
	"telkomsel-bot/util"
)

type Handler struct {
	bot         *gotgbot.Bot
	sessions    *model.SessionManager
	auth        *telkomsel.Auth
	api         *telkomsel.Client
	adminID     int64
	otpListener *otp.Listener

	otpChans   map[string]chan string // Key: FullPhone
	otpChansMu sync.Mutex

	pendingLogin   map[int64]string // Key: Telegram UserID, Value: FullPhone
	pendingLoginMu sync.Mutex

	autoStops   map[string]context.CancelFunc // Key: FullPhone
	autoStopsMu sync.Mutex
}

func NewHandler(bot *gotgbot.Bot, sessions *model.SessionManager, adminID int64, otpListener *otp.Listener) *Handler {
	return &Handler{
		bot:          bot,
		sessions:     sessions,
		auth:         telkomsel.NewAuth(),
		api:          telkomsel.NewClient(),
		adminID:      adminID,
		otpListener:  otpListener,
		otpChans:     make(map[string]chan string),
		pendingLogin: make(map[int64]string),
		autoStops:    make(map[string]context.CancelFunc),
	}
}

func (h *Handler) Register(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("start", h.handleStart))
	dispatcher.AddHandler(handlers.NewCallback(nil, h.handleCallback))
	dispatcher.AddHandler(handlers.NewMessage(message.All, h.handleMessage))
}

func (h *Handler) ValidateSessions() {
	telkomsel.ValidateSessions(h.sessions, h.api)
}

func (h *Handler) handleStart(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveSender.Id() != h.adminID {
		return nil
	}

	accounts := h.sessions.List()

	if len(accounts) == 0 {
		text := `🔰 *Telbot*

Selamat datang! Bot ini membantu kamu:
• Login otomatis ke MyTelkomsel (Multi Akun)
• Cek profil & kuota
• Beli paket otomatis

Tekan tombol di bawah untuk menambahkan akun pertama.`
		_, err := ctx.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kbLogin(),
		})
		return err
	}

	text := "🔰 *Telbot*\n\nPilih akun untuk dikelola atau tambahkan akun baru:"
	_, err := ctx.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kbAccounts(accounts),
	})
	return err
}

func (h *Handler) handleCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	cq := ctx.CallbackQuery
	if cq.From.Id != h.adminID {
		return nil
	}
	_, _ = cq.Answer(b, nil)

	chatID := cq.Message.GetChat().Id
	msgID := cq.Message.GetMessageId()
	userID := cq.From.Id
	data := cq.Data

	if data != "login" && data != "add_account" && data != "change_account" && data != "logout" {
		session := h.sessions.GetActive(userID)
		if session != nil && session.IsAwaiting() {
			session.State = model.StateLoggedIn
			h.sessions.Set(session.FullPhone, session)
		}
	}

	switch {
	case strings.HasPrefix(data, "select_acc_"):
		phone := strings.TrimPrefix(data, "select_acc_")
		h.sessions.SetActive(userID, phone)
		h.cbShowProfile(b, chatID, msgID, userID)

	case data == "login", data == "add_account":
		h.cbLogin(b, chatID, msgID, userID)

	case data == "change_account":
		accounts := h.sessions.List()
		kb := kbAccounts(accounts)
		h.editMsg(b, chatID, msgID, "🔰 *Telbot*\n\nPilih akun:", &kb)

	case data == "back_profile" || data == "refresh":
		h.cbShowProfile(b, chatID, msgID, userID)

	case data == "buy":
		h.cbShowMenu(b, chatID, msgID, userID)

	case data == "check_quota":
		h.cbCheckQuota(b, chatID, msgID, userID)

	case data == "pkg_ilmupedia":
		h.cbShowPayment(b, chatID, msgID, userID, "")

	case strings.HasPrefix(data, "pkg_offer_"):
		offerID := strings.TrimPrefix(data, "pkg_offer_")
		h.cbShowPayment(b, chatID, msgID, userID, offerID)

	case data == "pkg_custom":
		h.editMsg(b, chatID, msgID, "🆔 Kirim Offer ID paket yang ingin dibeli. \n\n Buka: https://my.telkomsel.com/web\n\n Contoh paket tiktok: https://my.telkomsel.com/app/package-details/bbc8df8c82679d736a792a39b7009499 \n\nAmbil ID Contoh: `bbc8df8c82679d736a792a39b7009499`", nil)
		session := h.sessions.GetActive(userID)
		if session != nil {
			session.State = model.StateAwaitingOfferID
			h.sessions.Set(session.FullPhone, session)
		}

	case strings.HasPrefix(data, "pay_qris"):
		offerID := strings.TrimPrefix(data, "pay_qris_")
		if offerID == "pay_qris" {
			offerID = ""
		}
		h.cbBuy(b, chatID, msgID, userID, "qris", offerID)

	case strings.HasPrefix(data, "pay_pulsa"):
		offerID := strings.TrimPrefix(data, "pay_pulsa_")
		if offerID == "pay_pulsa" {
			offerID = ""
		}
		h.cbBuy(b, chatID, msgID, userID, "AIRTIME", offerID)

	case data == "confirm_buy":
		h.cbConfirmBuy(b, chatID, msgID, userID)

	case data == "auto_buy":
		h.cbShowAutoMonitor(b, chatID, msgID, userID)

	case data == "auto_20":
		h.cbSetAutoInterval(b, chatID, msgID, userID, 20)

	case data == "auto_50":
		h.cbSetAutoInterval(b, chatID, msgID, userID, 50)

	case data == "auto_custom":
		h.editMsg(b, chatID, msgID, "⌨️ Kirim waktu monitor dalam menit.\n\nContoh: `30`", nil)
		session := h.sessions.GetActive(userID)
		if session != nil {
			session.State = model.StateAwaitingAutoInt
			h.sessions.Set(session.FullPhone, session)
		}

	case data == "auto_thresh_0":
		h.cbSetAutoThreshold(b, chatID, msgID, userID, 0)

	case data == "auto_thresh_100":
		h.cbSetAutoThreshold(b, chatID, msgID, userID, 100)

	case data == "auto_thresh_200":
		h.cbSetAutoThreshold(b, chatID, msgID, userID, 200)

	case data == "auto_thresh_custom":
		h.editMsg(b, chatID, msgID, "⌨️ Kirim batas minimum kuota dalam MB.\n\nContoh: `500`", nil)
		session := h.sessions.GetActive(userID)
		if session != nil {
			session.State = model.StateAwaitingAutoThreshold
			h.sessions.Set(session.FullPhone, session)
		}

	case data == "auto_pkg_ilmupedia":
		h.cbSetAutoPackage(b, chatID, msgID, userID, "ilmupedia")

	case strings.HasPrefix(data, "auto_pkg_offer_"):
		offerID := strings.TrimPrefix(data, "auto_pkg_offer_")
		h.cbSetAutoPackage(b, chatID, msgID, userID, offerID)

	case data == "auto_pkg_custom":
		h.editMsg(b, chatID, msgID, "🆔 Kirim Offer ID paket untuk auto-buy.\n\nContoh: `0fc00fd41bcd26376d806925d746705e`", nil)
		session := h.sessions.GetActive(userID)
		if session != nil {
			session.State = model.StateAwaitingAutoOffer
			h.sessions.Set(session.FullPhone, session)
		}

	case data == "auto_pay_pulsa":
		h.cbStartAutoBuy(b, chatID, msgID, userID)

	case data == "auto_stop":
		h.cbStopAutoBuy(b, chatID, msgID, userID)

	case data == "logout":
		session := h.sessions.GetActive(userID)
		if session != nil {
			h.stopAutoBuy(session.FullPhone)
			h.sessions.Delete(session.FullPhone)
		}
		accounts := h.sessions.List()
		if len(accounts) > 0 {
			kb := kbAccounts(accounts)
			h.editMsg(b, chatID, msgID, "👋 Sudah logout.\n\nPilih akun lain:", &kb)
		} else {
			kb := kbLogin()
			h.editMsg(b, chatID, msgID, "👋 Sudah logout.", &kb)
		}
	}

	return nil
}

func (h *Handler) handleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveSender.Id()
	if userID != h.adminID {
		return nil
	}

	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	if text == "" || strings.HasPrefix(text, "/") {
		return nil
	}

	h.pendingLoginMu.Lock()
	full := h.pendingLogin[userID]
	h.pendingLoginMu.Unlock()

	h.otpChansMu.Lock()
	_, hasOTP := h.otpChans[full]
	h.otpChansMu.Unlock()

	if hasOTP && full != "" {
		return h.handleOTPInput(b, ctx, userID, text)
	}

	session := h.sessions.GetActive(userID)

	if session != nil && session.State == model.StateAwaitingOfferID {
		session.State = model.StateIdle
		h.sessions.Set(session.FullPhone, session)
		kb := kbPaymentSelect(text)
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("🆔 Offer ID: `%s`\n\n💳 Pembayaran Via:", text), &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	if session != nil && session.State == model.StateAwaitingAutoInt {
		session.State = model.StateIdle
		minutes, parseErr := strconv.Atoi(text)
		if parseErr != nil || minutes <= 0 {
			_, err := ctx.EffectiveMessage.Reply(b, "❌ Masukkan angka yang valid (dalam menit).\n\nContoh: `30`", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
			return err
		}
		session.AutoBuyInterval = minutes
		h.sessions.Set(session.FullPhone, session)

		kb := kbAutoThreshold()
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("✅ Interval: *%d menit*\n\n📉 Pilih batas minimum kuota untuk auto-buy:", minutes), &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	if session != nil && session.State == model.StateAwaitingAutoThreshold {
		session.State = model.StateIdle
		mb, parseErr := strconv.Atoi(text)
		if parseErr != nil || mb < 0 {
			_, err := ctx.EffectiveMessage.Reply(b, "❌ Masukkan angka yang valid (dalam MB).\n\nContoh: `500`", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
			return err
		}
		session.AutoBuyThreshold = mb
		h.sessions.Set(session.FullPhone, session)

		threshStr := fmt.Sprintf("< %d MB", mb)
		if mb == 0 {
			threshStr = "Habis (0 MB)"
		}

		apiCtx := context.Background()
		offers, _ := h.api.GetRecommendedOffers(apiCtx, session)
		kb := kbAutoPackage(offers)
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("✅ Interval: *%d menit*\n📉 Batas Kuota: *%s*\n\n📦 Pilih paket untuk auto-buy:", session.AutoBuyInterval, threshStr), &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	if session != nil && session.State == model.StateAwaitingAutoOffer {
		session.State = model.StateIdle
		session.AutoBuyPackage = text
		h.sessions.Set(session.FullPhone, session)

		threshStr := fmt.Sprintf("< %d MB", session.AutoBuyThreshold)
		if session.AutoBuyThreshold == 0 {
			threshStr = "Habis (0 MB)"
		}

		kb := kbAutoPay()
		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("✅ Interval: *%d menit*\n📉 Batas Kuota: *%s*\n📦 Paket: *%s*\n\n💳 Pembayaran via:", session.AutoBuyInterval, threshStr, text), &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	if session != nil && session.State == model.StateAwaitingPhone {
		return h.handlePhoneInput(b, ctx, userID, text)
	}

	if util.IsPhoneNumber(text) {
		return h.handlePhoneInput(b, ctx, userID, text)
	}

	return nil
}