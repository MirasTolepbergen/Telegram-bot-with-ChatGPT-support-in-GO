package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// ChatGPT API base URL
const chatGPTAPIURL = "https://api.openai.com/v1/chat/completions"

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	openaiApiKey := os.Getenv("OPENAI_API_KEY")

	if telegramBotToken == "" || openaiApiKey == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN and OPENAI_API_KEY must be set")
	}

	// Create a new Telegram Bot instance
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			// Call ChatGPT API to get a response
			chatGPTResponse, err := getChatGPTResponse(update.Message.Text, openaiApiKey)
			if err != nil {
				msg.Text = "Sorry, something went wrong."
			} else {
				msg.Text = chatGPTResponse
			}

			bot.Send(msg)
		}
	}
}

// getChatGPTResponse sends a query to the ChatGPT API and returns the response
func getChatGPTResponse(query, apiKey string) (string, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model": "gpt-4",
			"messages": []map[string]string{
				{"role": "user", "content": query},
			},
		}).
		Post(chatGPTAPIURL)

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != 200 {
		return "", errors.New("non-200 status code from ChatGPT API")
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}
