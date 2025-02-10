package bot

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type telegramBot struct {
	client          *tgbotapi.BotAPI
	chatID          int64
	channelUsername string
}

var TeleBot *telegramBot = &telegramBot{}

func InitTeleBot(botToken string, chatId int64, channelUsername string) {
	for {
		bot, err := NewTelegramBot(botToken, chatId, channelUsername)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		TeleBot = bot
		break
	}

	// go func() {
	// 	// Get updates from Telegram
	// 	for {
	// 		u := tgbotapi.NewUpdate(0)
	// 		u.Timeout = 60
	// 		updates, err := TeleBot.client.GetUpdates(u)
	// 		if err != nil {
	// 			zap.L().With(zap.Error(err)).Error("get chat updates failed")
	// 			time.Sleep(1 * time.Second)
	// 		}
	// 		if len(updates) == 0 {
	// 			time.Sleep(5 * time.Second)
	// 			continue
	// 		}
	// 		needSleep := true
	// 		// Process incoming updates
	// 		for _, update := range updates {
	// 			u.Offset = update.UpdateID + 1
	// 			// Check if there is a command
	// 			if update.Message == nil {
	// 				continue
	// 			}
	// 			needSleep = false
	// 			switch update.Message.Command() {
	// 			case "pdf":
	// 				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the bot!")
	// 				TeleBot.client.Send(msg)
	// 			}
	// 		}
	// 		if needSleep {
	// 			time.Sleep(5 * time.Second)
	// 		}
	// 	}
	// }()

	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 20 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), "ws://103.82.133.178:9222")

	// defer cancel()

	// capture pdf

	b.Handle("/pdf", func(c tele.Context) error {
		data := c.Data()
		zap.L().With(zap.Any("data", data)).Info("creating pdf")
		ctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(zap.L().Sugar().Debugf))

		// task := chromedp.Tasks{
		// 	chromedp.Navigate(data),
		// 	// chromedp.Evaluate(script, nil),
		// 	// chromedp.WaitVisible(`body > footer`),
		// }
		// chromedp.Run(ctx, task)
		// slowScrollToEnd(ctx)
		var buf []byte
		data = "https://substack.com/redirect/2/eyJlIjoiaHR0cHM6Ly92aWV0aHVzdGxlci5zdWJzdGFjay5jb20vcC9tYXJrZXQtMDEyNy1uZ2F5LWJvbmctYm9uZy1haS12bz91dG1fY2FtcGFpZ249ZW1haWwtcG9zdCZyPWhqbXhhJnRva2VuPWV5SjFjMlZ5WDJsa0lqb3lPVFEyT1RZME5pd2ljRzl6ZEY5cFpDSTZNVFUxT0RVek16UXpMQ0pwWVhRaU9qRTNNemd3TURNNU5UVXNJbVY0Y0NJNk1UYzBNRFU1TlRrMU5Td2lhWE56SWpvaWNIVmlMVFk1TVRRNU5TSXNJbk4xWWlJNkluQnZjM1F0Y21WaFkzUnBiMjRpZlEuMGRCaEVTWmRwMjVTX2JoMy1sSG1hVGliaHk5eXZ6SVp1WjVQbnpKNnA4QSIsInAiOjE1NTg1MzM0MywicyI6NjkxNDk1LCJmIjpmYWxzZSwidSI6Mjk0Njk2NDYsImlhdCI6MTczODAwMzk1NSwiZXhwIjoxNzQwNTk1OTU1LCJpc3MiOiJwdWItMCIsInN1YiI6ImxpbmstcmVkaXJlY3QifQ.k_TScLohlTjXAhdufQfgY036vJqOD2CHO5Zv1pbqnV4"
		if err := chromedp.Run(ctx, printToPDF(data, &buf)); err != nil {
			zap.L().Error(err.Error())
		}
		zap.L().With(zap.Any("data", c.Data())).Info("created pdf")
		a := &tele.Document{File: tele.File{
			FileReader: bytes.NewReader(buf),
		},
			MIME: "application/pdf",
		}
		// b.Send(c.Sender(), a)
		// return nil
		return c.Send(a)

	})

	go b.Start()

	zap.L().With(zap.Int64("chat id", chatId)).Info("init telebot ok")
}

// print a specific pdf page.
func printToPDF(urlstr string, res *[]byte) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.Navigate(urlstr),

		// chromedp.Sleep(10 * time.Second),

		chromedp.Click(`div.main-content`),

		// if we send keys to `div.main-content`, it will return "Element is not focusable"
		// so we send keys to `body` which will not return error

		// if we use kb.End instead kb.PageDown, the result is same
		chromedp.SendKeys(`body`, kb.PageDown),
		chromedp.Sleep(1 * time.Second),

		chromedp.SendKeys(`body`, kb.PageDown),
		chromedp.Sleep(1 * time.Second),
		chromedp.SendKeys(`body`, kb.PageDown),
		chromedp.Sleep(1 * time.Second),
		chromedp.SendKeys(`body`, kb.PageDown),
		chromedp.Sleep(1 * time.Second),
		chromedp.SendKeys(`body`, kb.PageDown),
		chromedp.Sleep(1 * time.Second),
		chromedp.SendKeys(`body`, kb.PageDown),
		chromedp.Sleep(1 * time.Second),
		chromedp.SendKeys(`body`, kb.PageDown),
		chromedp.Sleep(1 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}

func NewTelegramBot(botToken string, chatId int64, channelUsername string) (*telegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		zap.L().With(zap.String("bot token", botToken)).With(zap.Error(err)).Error("init bot failed")
		return nil, err
	}

	// Set bot in debug mode (optional)
	bot.Debug = true

	// Set your chat ID here
	chatID := int64(chatId) // e.g., -123456789

	// Create a new message to send as notification
	if err != nil {
		zap.L().With(zap.String("bot token", botToken)).With(zap.Int64("chat id", chatID)).With(zap.Error(err)).Error("init bot failed")
		return nil, err
	}
	t := &telegramBot{
		client:          bot,
		chatID:          chatID,
		channelUsername: channelUsername,
	}
	return t, nil
}

func (t *telegramBot) SendMsg(msg string) {
	if (t.client) == nil {
		return
	}
	teleMsg := tgbotapi.NewMessage(t.chatID, msg)
	teleMsg.ParseMode = "HTML"
	_, err := t.client.Send(teleMsg)
	if err != nil {
		zap.L().With(zap.Error(err)).With(zap.Int64("chat id", t.chatID)).Error("send msg tele failed")
	}
}

func (t *telegramBot) SendMsgChannel(msg string) {
	if (t.client) == nil {
		return
	}
	teleMsg := tgbotapi.NewMessageToChannel(t.channelUsername, msg)
	teleMsg.ParseMode = "HTML"
	_, err := t.client.Send(teleMsg)
	if err != nil {
		zap.L().With(zap.Error(err)).With(zap.String("chat username", t.channelUsername)).Error("send msg tele failed")
	}
}

// slowScrollToEnd slowly scrolls to the end of the page
func slowScrollToEnd(ctx context.Context) error {
	var scrollHeight, currentHeight int64

	// Continuously scroll and wait
	for {
		zap.L().With(zap.Int("height", int(scrollHeight))).Info("h")
		// Get the current scroll height and page height

		chromedp.Evaluate(`document.body.scrollHeight`, &scrollHeight)
		chromedp.Evaluate(`window.scrollY`, &currentHeight)

		zap.L().With(zap.Int("height", int(scrollHeight))).Info("h")

		// Check if we are at the bottom
		if currentHeight+100 >= scrollHeight {
			// If we're close enough to the bottom, break the loop
			break
		}

		// Scroll a little bit more
		err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.scrollTo(0, window.scrollY + 300)`, nil), // Scroll 300px down
		)
		if err != nil {
			return err
		}

		// Wait a bit before scrolling again
		time.Sleep(100 * time.Millisecond) // Adjust this for slower scrolling speed
	}

	return nil
}
