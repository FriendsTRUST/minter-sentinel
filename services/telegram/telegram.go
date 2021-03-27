package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

func New(token string) (*tgbotapi.BotAPI, error) {
	return tgbotapi.NewBotAPI(token)
}
