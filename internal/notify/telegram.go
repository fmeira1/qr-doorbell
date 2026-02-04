package notify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type TelegramNotifier struct {
	token  string
	chatID string
}

func NewTelegramNotifier(token, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		token:  token,
		chatID: chatID,
	}
}

func (t *TelegramNotifier) Notify(text string) error {
	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		t.token,
	)

	data := url.Values{}
	data.Set("chat_id", t.chatID)
	data.Set("text", text)

	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// NotifyWithButtons envía el mensaje con botones: uno para abrir el chat (URL) y dos respuestas rápidas (callback).
func (t *TelegramNotifier) NotifyWithButtons(text, sessionID, chatURL string) error {
	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		t.token,
	)

	// Primera fila: botón que abre el link del chat. Segunda fila: respuestas rápidas por callback.
	inlineKeyboard := [][]map[string]string{
		{{"text": "🔗 Abrir chat", "url": chatURL}},
		{
			{"text": "✅ Hola! Ya voy", "callback_data": "reply:coming:" + sessionID},
			{"text": "❌ No puedo atender ahora", "callback_data": "reply:busy:" + sessionID},
		},
	}
	replyMarkupBytes, _ := json.Marshal(map[string]interface{}{"inline_keyboard": inlineKeyboard})
	replyMarkup := string(replyMarkupBytes)

	data := url.Values{}
	data.Set("chat_id", t.chatID)
	data.Set("text", text)
	data.Set("reply_markup", replyMarkup)

	_, err := http.PostForm(endpoint, data)
	return err
}
