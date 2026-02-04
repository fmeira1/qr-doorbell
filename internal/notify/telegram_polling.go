package notify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type TelegramUpdate struct {
	UpdateID      int `json:"update_id"`
	CallbackQuery *struct {
		Data string `json:"data"`
	} `json:"callback_query"`
}

type TelegramResponse struct {
	Result []TelegramUpdate `json:"result"`
}

func (t *TelegramNotifier) StartPolling(onReply func(sessionID, text string)) {
	var offset int

	go func() {
		for {
			url := fmt.Sprintf(
				"https://api.telegram.org/bot%s/getUpdates?timeout=10&offset=%d",
				t.token,
				offset,
			)

			resp, err := http.Get(url)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			var result TelegramResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				time.Sleep(2 * time.Second)
				continue
			}
			resp.Body.Close()

			for _, update := range result.Result {
				offset = update.UpdateID + 1

				if update.CallbackQuery == nil {
					continue
				}

				parts := strings.Split(update.CallbackQuery.Data, ":")
				if len(parts) != 3 {
					continue
				}

				action, kind, sessionID := parts[0], parts[1], parts[2]
				if action != "reply" {
					continue
				}

				var text string
				switch kind {
				case "coming":
					text = "Hola! Ya voy 🙂"
				case "busy":
					text = "Hola! En este momento no puedo atender."
				default:
					continue
				}

				onReply(sessionID, text)
			}

			time.Sleep(1 * time.Second)
		}
	}()
}
