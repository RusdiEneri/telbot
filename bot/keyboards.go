package bot

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"telkomsel-bot/model"
	"telkomsel-bot/telkomsel"
)

func kbAccounts(accounts []*model.Session) gotgbot.InlineKeyboardMarkup {
	var rows [][]gotgbot.InlineKeyboardButton
	for _, acc := range accounts {
		phone := acc.FullPhone
		if phone == "" {
			phone = acc.Phone
		}
		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: fmt.Sprintf("📱 +%s", phone), CallbackData: "select_acc_" + phone},
		})
	}
	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "➕ Tambah Akun", CallbackData: "add_account"},
	})
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func kbMenuWithOffers(offers []telkomsel.RecommendedOffer) gotgbot.InlineKeyboardMarkup {
	var rows [][]gotgbot.InlineKeyboardButton

	for _, o := range offers {
		label := fmt.Sprintf("📦 %s - Rp%s", o.Name, o.Price)
		if len(label) > 64 {
			label = label[:61] + "..."
		}
		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: label, CallbackData: "pkg_offer_" + o.ID},
		})
	}

	rows = append(rows,
		[]gotgbot.InlineKeyboardButton{
			{Text: "🆔 Custom Id", CallbackData: "pkg_custom"},
		},
		[]gotgbot.InlineKeyboardButton{
			{Text: "🤖 Beli Otomatis", CallbackData: "auto_buy"},
		},
		[]gotgbot.InlineKeyboardButton{
			{Text: "🔙 Kembali", CallbackData: "back_profile"},
		},
	)

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func kbLogin() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: "➕ Tambah Akun", CallbackData: "add_account"}},
		},
	}
}

func kbProfile(accountsCount int) gotgbot.InlineKeyboardMarkup {
	rows := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "🛒 Beli Paket", CallbackData: "buy"},
			{Text: "📊 Cek Kouta", CallbackData: "check_quota"},
		},
		{{Text: "🔄 Refresh", CallbackData: "refresh"}},
	}

	if accountsCount > 1 {
		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: "👥 Ganti Akun", CallbackData: "change_account"},
		})
	}

	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "👋 Logout", CallbackData: "logout"},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func kbMenu() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🛒 Beli Paket", CallbackData: "buy"},
				{Text: "📊 Cek Kuota", CallbackData: "check_quota"},
			},
			{
				{Text: "🔙 Kembali", CallbackData: "back_profile"},
			},
		},
	}
}

func kbBack(backData string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: "🔙 Kembali", CallbackData: backData}},
		},
	}
}

func kbPaymentSelect(offerID string) gotgbot.InlineKeyboardMarkup {
	payPulsa := "pay_pulsa"
	payQris := "pay_qris"
	if offerID != "" {
		payPulsa = "pay_pulsa_" + offerID
		payQris = "pay_qris_" + offerID
	}
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "💰 Pulsa", CallbackData: payPulsa},
				{Text: "📱 QRIS", CallbackData: payQris},
			},
			{
				{Text: "🔙 Kembali", CallbackData: "buy"},
			},
		},
	}
}

func kbConfirmBuy() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "✅ Ya, Beli Paket Ini", CallbackData: "confirm_buy"},
			},
			{
				{Text: "❌ Batal", CallbackData: "buy"},
			},
		},
	}
}

func kbAutoRunning() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🛑 Stop Auto-Buy", CallbackData: "auto_stop"},
			},
			{
				{Text: "🔙 Kembali", CallbackData: "back_profile"},
			},
		},
	}
}

func kbAutoMonitor() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "⏱️ 20 Menit", CallbackData: "auto_20"},
				{Text: "⏱️ 50 Menit", CallbackData: "auto_50"},
			},
			{
				{Text: "⌨️ Custom", CallbackData: "auto_custom"},
			},
			{
				{Text: "🔙 Kembali", CallbackData: "back_profile"},
			},
		},
	}
}

func kbAutoThreshold() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📉 0 MB (Habis)", CallbackData: "auto_thresh_0"},
				{Text: "📉 100 MB", CallbackData: "auto_thresh_100"},
			},
			{
				{Text: "📉 200 MB", CallbackData: "auto_thresh_200"},
				{Text: "⌨️ Custom", CallbackData: "auto_thresh_custom"},
			},
			{
				{Text: "🔙 Kembali", CallbackData: "back_profile"},
			},
		},
	}
}

func kbAutoPackage(offers []telkomsel.RecommendedOffer) gotgbot.InlineKeyboardMarkup {
	var rows [][]gotgbot.InlineKeyboardButton
	for _, o := range offers {
		label := fmt.Sprintf("📦 %s - Rp%s", o.Name, o.Price)
		if len(label) > 64 {
			label = label[:61] + "..."
		}
		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: label, CallbackData: "auto_pkg_offer_" + o.ID},
		})
	}
	rows = append(rows,
		[]gotgbot.InlineKeyboardButton{
			{Text: "🆔 Custom Offer ID", CallbackData: "auto_pkg_custom"},
		},
		[]gotgbot.InlineKeyboardButton{
			{Text: "🔙 Kembali", CallbackData: "back_profile"},
		},
	)
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func kbAutoPay() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "💰 Pulsa", CallbackData: "auto_pay_pulsa"},
			},
			{
				{Text: "🔙 Kembali", CallbackData: "back_profile"},
			},
		},
	}
}