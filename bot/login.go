package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"telkomsel-bot/telkomsel"
	"telkomsel-bot/util"
)

func (h *Handler) handlePhoneInput(b *gotgbot.Bot, ctx *ext.Context, userID int64, input string) error {
	local, full, err := util.ValidatePhone(input)
	if err != nil {
		_, replyErr := ctx.EffectiveMessage.Reply(b, "❌ Nomor tidak valid. Contoh: `812xxxxxxxx`", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
		return replyErr
	}

	_, _ = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("📱 Login: +%s\n\n🔄 Proses login...", full), nil)

	otpChan := make(chan string, 1)
	h.otpChansMu.Lock()
	h.otpChans[full] = otpChan
	h.otpChansMu.Unlock()

	h.pendingLoginMu.Lock()
	h.pendingLogin[userID] = full
	h.pendingLoginMu.Unlock()

	defer func() {
		h.otpChansMu.Lock()
		delete(h.otpChans, full)
		h.otpChansMu.Unlock()

		h.pendingLoginMu.Lock()
		delete(h.pendingLogin, userID)
		h.pendingLoginMu.Unlock()
	}()

	apiCtx := context.Background()

	otpCallback := func() (string, error) {
		_, _ = ctx.EffectiveMessage.Reply(b, "📲 OTP dikirim ke HP kamu.\n\n🔢 *Kirim kode OTP:*", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})

		if h.otpListener != nil {
			waitCtx, cancel := context.WithTimeout(apiCtx, 3*time.Minute)
			defer cancel()

			webhookChan := make(chan string, 1)
			go func() {
				otp, err := h.otpListener.WaitForOTP(waitCtx, local)
				if err == nil && otp != "" {
					webhookChan <- otp
				}
			}()

			select {
			case otp := <-otpChan:
				return otp, nil
			case otp := <-webhookChan:
				_, _ = ctx.EffectiveMessage.Reply(b, "🤖 *Auto OTP diterima dari SMS Forwarder!* Memverifikasi...", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
				return otp, nil
			case <-time.After(2 * time.Minute):
				return "", fmt.Errorf("OTP timeout (2 menit)")
			}
		}

		select {
		case otp := <-otpChan:
			return otp, nil
		case <-time.After(2 * time.Minute):
			return "", fmt.Errorf("OTP timeout (2 menit)")
		}
	}

	session, loginErr := h.auth.Login(apiCtx, local, otpCallback)
	if loginErr != nil {
		h.sessions.Delete(full)
		log.Printf("[Login] Error for account %s: %v", full, loginErr)
		_, replyErr := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("❌ Login gagal: %s", loginErr.Error()), &gotgbot.SendMessageOpts{
			ReplyMarkup: kbLogin(),
		})
		return replyErr
	}

	// Paksa FullPhone dan Phone sesuai hasil parsing
	session.FullPhone = full
	session.Phone = local

	existing := h.sessions.Get(full)
	if existing != nil {
		session.AutoBuyInterval = existing.AutoBuyInterval
		session.AutoBuyPackage = existing.AutoBuyPackage
		session.AutoBuyPayment = existing.AutoBuyPayment
		session.AutoBuyActive = existing.AutoBuyActive
		session.PendingOfferID = existing.PendingOfferID
		session.PendingPayment = existing.PendingPayment
	}
    
	h.sessions.Set(full, session)
	h.sessions.SetActive(userID, full) // Otomatis switch ke akun baru

	profile, profileErr := h.api.GetFullProfile(context.Background(), session)
	var profileText string
	if profileErr == nil && profile != nil {
		profileText = "\n" + telkomsel.FormatProfile(profile)
	}

	kb := kbProfile(len(h.sessions.List()))
	_, replyErr := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("✅ Login berhasil!%s\nPilih aksi:", profileText), &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	return replyErr
}

func (h *Handler) handleOTPInput(b *gotgbot.Bot, ctx *ext.Context, userID int64, text string) error {
	h.pendingLoginMu.Lock()
	full := h.pendingLogin[userID]
	h.pendingLoginMu.Unlock()

	h.otpChansMu.Lock()
	otpChan, hasOTP := h.otpChans[full]
	h.otpChansMu.Unlock()

	if !hasOTP {
		return nil
	}

	select {
	case otpChan <- text:
		_, _ = ctx.EffectiveMessage.Reply(b, "✓ OTP diterima, memproses...", nil)
	default:
		_, _ = ctx.EffectiveMessage.Reply(b, "⚠️ Channel penuh, coba login ulang.", nil)
	}
	return nil
}

func (h *Handler) cbLogin(b *gotgbot.Bot, chatID, msgID, userID int64) {
	h.editMsg(b, chatID, msgID, "📱 Kirim nomor HP Telkomsel kamu.\n\nContoh: `812xxxxxxxx`", nil)
}