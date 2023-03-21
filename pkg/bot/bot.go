package bot

import (
	"time"

	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

const defaultEnvFile = ".env"
const defaultTimeout = 10 * time.Second

type BotConfig struct {
	// Telegram bot token
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN" required:"true"`

	// OpenAI API key
	OpenAIAPIKey string `env:"OPENAI_API_KEY" required:"true"`

	// Valid chat ID (whitelist)
	ValidChatID []int64 `env:"VALID_CHAT_ID"`
}

func Execute() {
	err := godotenv.Load(defaultEnvFile)
	if err != nil {
		log.Warnf("failed to load .env file: %+v", err)
	}

	var config BotConfig
	if err := env.Set(&config); err != nil {
		log.Fatal(err)
	}

	pref := tele.Settings{
		Token:  config.TelegramBotToken,
		Poller: &tele.LongPoller{Timeout: defaultTimeout},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(config.ValidChatID) > 0 {
		bot.Use(whitelist(config.ValidChatID...))
	}

	bot.Use(responseTimer)
	bot.Use(messageLogger)

	chatGPTService := NewChatGPTService(config.OpenAIAPIKey)

	bot.Handle("/gpt", chatGPTService.HandleNewChat)
	bot.Handle(tele.OnText, chatGPTService.HandleTextReply)
	bot.Handle("/set", chatGPTService.HandleSetCommand)
	bot.Handle("/temperature", chatGPTService.HandleTemperatureCommand)
	bot.Handle("/help", HandleHelpCommand)
	bot.Handle("/tc", chatGPTService.HandleTCCommand)

	log.Infof("Starting bot")
	bot.Start()
}
