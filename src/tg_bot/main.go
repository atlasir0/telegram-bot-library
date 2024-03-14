package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const (
	BotToken   = "7145361114:AAGcDmLWHv9eyeQTjcj1djRA1oDcCJBmuKg"
	WebhookURL = "https://df4b-178-207-154-253.ngrok-free.app"
)

var (
	messageMapMutex sync.RWMutex
)

func startListening(bot *tgbotapi.BotAPI) {
	updates := bot.ListenForWebhook("/")
	go http.ListenAndServe(":8080", nil)
	fmt.Println("start listen :8080")

	for update := range updates {
		go handleMessage(bot, update)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	key := strings.ToLower(update.Message.Text)
	if key == "/help" {
		var avCommands strings.Builder
		avCommands.WriteString("Available commands:\n")
		messageMapMutex.RLock()
		defer messageMapMutex.RUnlock()

		for cmd := range messageMap {
			avCommands.WriteString(cmd)
			avCommands.WriteString("\n")
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, avCommands.String())
		bot.Send(msg)
	} else {
		messageMapMutex.RLock()
		answer, ok := messageMap[key]
		messageMapMutex.RUnlock()

		if ok {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, answer)
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Command not found")
			bot.Send(msg)
		}
	}
}

var (
	messageMap = make(map[string]string)
)

func initializeBot() (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot, nil
}

func setWebhook(bot *tgbotapi.BotAPI) error {
	_, err := bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	return err
}

func start(commands []struct{ Command, Response string }) {
	for _, cmd := range commands {
		messageMap[cmd.Command] = cmd.Response
	}
	bot, err := initializeBot()
	if err != nil {
		log.Fatal(err)
	}
	err = setWebhook(bot)
	if err != nil {
		log.Fatal(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	startListening(bot)
}

func main() {
	commands := []struct {
		Command, Response string
	}{
		{"/add", "ins '/add hello:Hi there!'"},
		{"/cat", "https://s3.amazonaws.com/freecodecamp/running-cats.jpg"},
		{"/g", "seg"},
	}

	start(commands)
}
