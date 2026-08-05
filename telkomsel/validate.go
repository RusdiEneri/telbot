package telkomsel

import (
	"context"
	"errors"
	"log"

	"telkomsel-bot/model"
)

func ValidateSessions(sessions *model.SessionManager, api *Client) {
	all := sessions.List()
	if len(all) == 0 {
		return
	}

	log.Printf("🔍 Validating %d saved session(s)...", len(all))
	apiCtx := context.Background()

	for _, session := range all {
		if !session.IsLoggedIn() {
			log.Printf("  ⏭ Account +%s: not logged in (state=%s), skipping", session.FullPhone, session.State)
			continue
		}

		_, _, err := api.GetBalance(apiCtx, session)
		if err != nil {
			if errors.Is(err, ErrUnauthorized) {
				log.Printf("  ❌ Account +%s: token expired, removing session", session.FullPhone)
				sessions.Delete(session.FullPhone)
			} else {
				log.Printf("  ⚠️ Account +%s: API error (%v), keeping session", session.FullPhone, err)
			}
		} else {
			log.Printf("  ✅ Account +%s: session valid", session.FullPhone)
		}
	}
}