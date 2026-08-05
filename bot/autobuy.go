package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"telkomsel-bot/telkomsel"
	"telkomsel-bot/util"
)

func (h *Handler) cbShowAutoMonitor(b *gotgbot.Bot, chatID, msgID, userID int64) {
	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	if session.AutoBuyActive {
		kb := kbAutoRunning()
		threshStr := fmt.Sprintf("< %d MB", session.AutoBuyThreshold)
		if session.AutoBuyThreshold == 0 {
			threshStr = "Habis (0 MB)"
		}
		h.editMsg(b, chatID, msgID, fmt.Sprintf(
			"🤖 *Auto-Buy Sedang Aktif!*\n\n⏱ Interval: *%d menit*\n📉 Batas Kuota: *%s*\n📦 Paket: *%s*\n💳 Bayar: *Pulsa*\n\nMonitor berjalan di background...",
			session.AutoBuyInterval, threshStr, session.AutoBuyPackage,
		), &kb)
		return
	}

	kb := kbAutoMonitor()
	h.editMsg(b, chatID, msgID, "⏱ Masukan waktu monitor untuk mengecek sisa kuota atau masa aktif kuota:", &kb)
}

func (h *Handler) cbSetAutoInterval(b *gotgbot.Bot, chatID, msgID, userID int64, minutes int) {
	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	session.AutoBuyInterval = minutes
	h.sessions.Set(session.FullPhone, session)

	kb := kbAutoThreshold()
	h.editMsg(b, chatID, msgID, fmt.Sprintf("✅ Interval: *%d menit*\n\n📉 Pilih batas minimum kuota untuk auto-buy:", minutes), &kb)
}

func (h *Handler) cbSetAutoThreshold(b *gotgbot.Bot, chatID, msgID, userID int64, threshold int) {
	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	session.AutoBuyThreshold = threshold
	h.sessions.Set(session.FullPhone, session)

	threshStr := fmt.Sprintf("< %d MB", threshold)
	if threshold == 0 {
		threshStr = "Habis (0 MB)"
	}

	h.editMsg(b, chatID, msgID, "⏳ Mengambil rekomendasi paket...", nil)

	apiCtx := context.Background()
	offers, _ := h.api.GetRecommendedOffers(apiCtx, session)

	kb := kbAutoPackage(offers)
	h.editMsg(b, chatID, msgID, fmt.Sprintf("✅ Interval: *%d menit*\n📉 Batas Kuota: *%s*\n\n📦 Pilih paket untuk auto-buy:", session.AutoBuyInterval, threshStr), &kb)
}

func (h *Handler) cbSetAutoPackage(b *gotgbot.Bot, chatID, msgID, userID int64, packageID string) {
	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	session.AutoBuyPackage = packageID
	h.sessions.Set(session.FullPhone, session)

	threshStr := fmt.Sprintf("< %d MB", session.AutoBuyThreshold)
	if session.AutoBuyThreshold == 0 {
		threshStr = "Habis (0 MB)"
	}

	kb := kbAutoPay()
	h.editMsg(b, chatID, msgID, fmt.Sprintf(
		"✅ Interval: *%d menit*\n📉 Batas Kuota: *%s*\n📦 Paket: *%s*\n\n💳 Pembayaran via:",
		session.AutoBuyInterval, threshStr, packageID,
	), &kb)
}

func (h *Handler) runAutoBuyMonitor(ctx context.Context, b *gotgbot.Bot, chatID int64, phone string) {
	session := h.sessions.Get(phone)
	if session == nil {
		return
	}

	interval := time.Duration(session.AutoBuyInterval) * time.Minute
	offerID := session.AutoBuyPackage
	if offerID == "ilmupedia" {
		offerID = ""
	}

	log.Printf("[AutoBuy] Started monitor for %s: every %d min, package=%s", phone, session.AutoBuyInterval, session.AutoBuyPackage)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[AutoBuy] Monitor stopped for %s", phone)
			return
		case <-time.After(interval):
		}

		session = h.sessions.Get(phone)
		if session == nil || !session.IsLoggedIn() || !session.AutoBuyActive {
			log.Printf("[AutoBuy] Session invalid, stopping monitor for %s", phone)
			return
		}

		apiCtx := context.Background()
		quota, err := h.api.CheckQuota(apiCtx, session)
		if err != nil {
			log.Printf("[AutoBuy] Failed to check quota for %s: %v", phone, err)
			continue
		}

		var needsBuy bool = true
		var matchedOrderID string
		var trackedItem *telkomsel.QuotaItem

		if session.AutoBuyOrderID != "" {
			for _, group := range quota.Groups {
				for _, item := range group.Items {
					if item.OrderID == session.AutoBuyOrderID {
						it := item
						trackedItem = &it
						matchedOrderID = item.OrderID
						break
					}
				}
				if trackedItem != nil {
					break
				}
			}
		}

		if trackedItem != nil {
			if util.ParseQuotaToMB(trackedItem.Remaining) >= float64(session.AutoBuyThreshold) {
				needsBuy = false
			}
		} else {
			var totalRemaining float64
			for _, group := range quota.Groups {
				for _, item := range group.Items {
					totalRemaining += util.ParseQuotaToMB(item.Remaining)
				}
			}
			if totalRemaining >= float64(session.AutoBuyThreshold) {
				needsBuy = false
			}
		}

		if !needsBuy {
			log.Printf("[AutoBuy] Quota OK for %s, skipping purchase", phone)
			continue
		}

		log.Printf("[AutoBuy] Quota depleted for %s, purchasing...", phone)
		_, _ = b.SendMessage(chatID, "🤖 *Auto-Buy:* Kuota habis terdeteksi! Membeli otomatis...", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})

		result, buyErr := h.api.BuyIlmupedia(apiCtx, session, offerID, "AIRTIME")
		if buyErr != nil {
			_, _ = b.SendMessage(chatID, fmt.Sprintf("❌ Auto-buy gagal: %s", buyErr.Error()), &gotgbot.SendMessageOpts{
				ReplyMarkup: kbAutoRunning(),
			})
			continue
		}

		if result.OrderID != "" && matchedOrderID == "" {
			session.AutoBuyOrderID = result.OrderID
			h.sessions.Set(session.FullPhone, session)
			log.Printf("[AutoBuy] Updated tracked OrderID to %s for %s", result.OrderID, phone)
		}

		_, _ = b.SendMessage(chatID, fmt.Sprintf("✅ *Auto-Buy Berhasil!*\n\n%s", telkomsel.FormatPurchaseResult(result, "AIRTIME")), &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kbAutoRunning(),
		})
	}
}

func (h *Handler) cbStartAutoBuy(b *gotgbot.Bot, chatID, msgID, userID int64) {
	session, ok := h.checkSession(b, chatID, msgID, userID)
	if !ok {
		return
	}

	if session.AutoBuyInterval <= 0 || session.AutoBuyPackage == "" {
		kb := kbAutoMonitor()
		h.editMsg(b, chatID, msgID, "⚠️ Konfigurasi belum lengkap. Mulai ulang.", &kb)
		return
	}

	session.AutoBuyPayment = "AIRTIME"
	session.AutoBuyActive = true
	h.sessions.Set(session.FullPhone, session)

	h.stopAutoBuy(session.FullPhone)

	autCtx, cancel := context.WithCancel(context.Background())
	h.autoStopsMu.Lock()
	h.autoStops[session.FullPhone] = cancel
	h.autoStopsMu.Unlock()

	threshStr := fmt.Sprintf("< %d MB", session.AutoBuyThreshold)
	if session.AutoBuyThreshold == 0 {
		threshStr = "Habis (0 MB)"
	}

	kb := kbAutoRunning()
	h.editMsg(b, chatID, msgID, fmt.Sprintf(
		"🤖 *Auto-Buy Aktif!*\n\n⏱ Interval: *%d menit*\n📉 Batas Kuota: *%s*\n📦 Paket: *%s*\n💳 Bayar: *Pulsa*\n\nMonitor berjalan di background...",
		session.AutoBuyInterval, threshStr, session.AutoBuyPackage,
	), &kb)

	go h.runAutoBuyMonitor(autCtx, b, chatID, session.FullPhone)
}

func (h *Handler) cbStopAutoBuy(b *gotgbot.Bot, chatID, msgID, userID int64) {
	session := h.sessions.GetActive(userID)
	if session == nil {
		return
	}

	h.stopAutoBuy(session.FullPhone)

	session.AutoBuyActive = false
	h.sessions.Set(session.FullPhone, session)

	kb := kbProfile(len(h.sessions.List()))
	h.editMsg(b, chatID, msgID, "🛑 Auto-buy dihentikan.", &kb)
}

func (h *Handler) stopAutoBuy(phone string) {
	h.autoStopsMu.Lock()
	if cancel, ok := h.autoStops[phone]; ok {
		cancel()
		delete(h.autoStops, phone)
	}
	h.autoStopsMu.Unlock()
}