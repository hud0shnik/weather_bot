package handler

import (
	"strings"

	"github.com/hud0shnik/weather_bot/internal/api"
	"github.com/hud0shnik/weather_bot/internal/commands"
	"github.com/hud0shnik/weather_bot/internal/repository"
	"github.com/hud0shnik/weather_bot/internal/send"
	"github.com/hud0shnik/weather_bot/internal/telegram"
)

// Функция обработки сообщений
func Respond(botUrl string, update telegram.Update) {

	// Обработчик команд
	if update.Message.Text != "" {

		request := append(strings.Split(update.Message.Text, " "), "", "")

		// Вывод реквеста для тестов
		// fmt.Println("request: \t", request)

		switch request[0] {
		case "/week":
			api.SendDailyWeather(botUrl, update.Message.Chat.ChatId, 7)
		case "/weather":
			api.SendClassicWeather(botUrl, update.Message.Chat.ChatId)
		case "/current":
			api.SendCurrentWeather(botUrl, update.Message.Chat.ChatId)
		case "/sun":
			api.SendSunInfo(botUrl, update.Message.Chat.ChatId)
		case "/set":
			repository.SetCoordinates(botUrl, update.Message.Chat.ChatId, request[1], request[2])
		case "/help", "/start":
			commands.Help(botUrl, update.Message.Chat.ChatId)
		default:
			send.SendMsg(botUrl, update.Message.Chat.ChatId, "Я не понимаю, воспользуйтесь /help")
		}

	} else {

		// Если пользователь отправил не сообщение:
		send.SendMsg(botUrl, update.Message.Chat.ChatId, "Пока я воспринимаю только текст")

	}

}
