package bot

import (
	"os"
	"time"

	"github.com/joho/godotenv"

	masker "github.com/ggwhite/go-masker"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

const defaultEnvFile = ".env"
const defaultTimeout = 10 * time.Second

func Execute() {
	err := godotenv.Load(defaultEnvFile)
	if err != nil {
		log.Warnf("failed to load .env file: %+v", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	log.Infof("TELEGRAM_BOT_TOKEN: %s", masker.Address(botToken))
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
		return
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	log.Infof("OPENAI_API_KEY: %s", masker.Address(openaiAPIKey))
	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
		return
	}

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
		bot.Use(whitelist(validChatID...))
	}

	bot.Use(responseTimer)
	bot.Use(messageLogger)

	chatGPT := NewChatGPT(openaiAPIKey)

	bot.Handle("/gpt", chatGPT.handleNewChat)
	bot.Handle(tele.OnText, chatGPT.handleReply)
	bot.Handle("/set", chatGPT.setSystemContent)
	bot.Handle("/temperature", chatGPT.setTemperature)

	log.Infof("Starting bot")
	bot.Start()
}
