package bot

import (
	"time"

	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"
	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/store"
	openai "github.com/sashabaranov/go-openai"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

const (
	defaultEnvFile = ".env"
	defaultTimeout = 10 * time.Second
	temperature    = 0.0
	maxTokens      = 0
)

type BotConfig struct {
	// Telegram bot token
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN" required:"true"`

	// OpenAI API key
	OpenAIAPIKey string `env:"OPENAI_API_KEY" required:"true"`

	// whitelist (chat ID)
	BotWhitelist []int64 `env:"BOT_WHITELIST"`

	// enable the /image command
	EnableImageCommand bool `env:"ENABLE_IMAGE_COMMAND"`
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

	if len(config.BotWhitelist) > 0 {
		bot.Use(whitelist(config.BotWhitelist...))
	}

	bot.Use(responseTimer)
	bot.Use(messageLogger)

	chatStore := store.New("chats")
	gpt4oService := ChatGPTService{
		Client:      openai.NewClient(config.OpenAIAPIKey),
		ChatStore:   chatStore,
		Model:       "gpt-4o-mini",
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	bot.Handle("/gpt", gpt4oService.HandleNewChat)
	bot.Handle(tele.OnText, gpt4oService.HandleTextReply)
	bot.Handle("/help", HandleHelpCommand)
	bot.Handle("/tc", gpt4oService.HandleTCCommand)
	bot.Handle("/en", gpt4oService.HandleENCommand)
	// bot.Handle("/jp", gpt4oService.HandleJPCommand)
	bot.Handle("/polish", gpt4oService.HandlePolishCommand)

	// o1 preview model
	o1Service := ChatGPTService{
		Client:      openai.NewClient(config.OpenAIAPIKey),
		ChatStore:   chatStore,
		Model:       "o1-preview",
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	bot.Handle("/o1", o1Service.HandleNewChat)

	if config.EnableImageCommand {
		log.Infof("enabling /image command")
		imageService := NewImageService(config.OpenAIAPIKey)
		bot.Handle("/image", imageService.HandleImageCommand)
	}

	log.Infof("Starting bot")
	bot.Start()
}
