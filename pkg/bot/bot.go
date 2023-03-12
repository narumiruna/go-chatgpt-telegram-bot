package bot

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

const defaultTimeout = 10 * time.Second

func Execute() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	validChatID, err := parseInt64(os.Getenv("VALID_CHAT_ID"))
	if err != nil {
		log.Fatalf("failed to parse VALID_CHAT_ID: %+v", err)
		return
	}

	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: defaultTimeout},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(validChatID) > 0 {
		bot.Use(middleware.Whitelist(validChatID...))
	}

	bot.Use(responseTime)

	chatGPT := NewChatGPT(openaiAPIKey, validChatID)

	bot.Handle("/gpt", chatGPT.newChat)
	bot.Handle(tele.OnText, chatGPT.reply)

	log.Infof("Starting bot")
	bot.Start()
}
