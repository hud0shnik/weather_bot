package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tgBot/mods"

	"github.com/spf13/viper"
)

// Структуры для работы с Telegram API

type telegramResponse struct {
	Result []update `json:"result"`
}

type update struct {
	UpdateId int     `json:"update_id"`
	Message  message `json:"message"`
}

type message struct {
	Chat    chat    `json:"chat"`
	Text    string  `json:"text"`
	Sticker sticker `json:"sticker"`
}

type chat struct {
	ChatId int `json:"id"`
}

type sticker struct {
	File_id string `json:"file_id"`
}

func main() {

	// Инициализация конфига (токенов)
	err := initConfig()
	if err != nil {
		log.Fatalf("initConfig error: %s", err)
		return
	}

	// Url бота для отправки и приёма сообщений
	botUrl := "https://api.telegram.org/bot" + viper.GetString("token")
	offSet := 0

	// Цикл работы бота
	for {

		// Получение апдейтов
		updates, err := getUpdates(botUrl, offSet)
		if err != nil {
			log.Fatalf("getUpdates error: %s", err)
			return
		}

		// Обработка апдейтов
		for _, update := range updates {
			respond(botUrl, update)
			offSet = update.UpdateId + 1
		}

		// Вывод в консоль для тестов
		// fmt.Println(updates)
	}
}

// Функция получения апдейтов
func getUpdates(botUrl string, offset int) ([]update, error) {

	// Rest запрос для получения апдейтов
	resp, err := http.Get(botUrl + "/getUpdates?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Запись и обработка полученных данных
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var restResponse telegramResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}

	return restResponse.Result, nil
}

// Функция обработки сообщений
func respond(botUrl string, update update) {

	// Обработчик команд
	if update.Message.Text != "" {

		request := append(strings.Split(update.Message.Text, " "), "", "")

		// Вывод реквеста для тестов
		// fmt.Println("request: \t", request)

		switch request[0] {
		case "/week":
			mods.SendDailyWeather(botUrl, update.Message.Chat.ChatId, 7)
		case "/weather":
			mods.SendCurrentWeather(botUrl, update.Message.Chat.ChatId)
			mods.SendDailyWeather(botUrl, update.Message.Chat.ChatId, 2)
		case "/current":
			mods.SendCurrentWeather(botUrl, update.Message.Chat.ChatId)
		case "/sun":
			mods.SendSunInfo(botUrl, update.Message.Chat.ChatId)
		case "/set":
			mods.SetPlace(botUrl, update.Message.Chat.ChatId, request[1], request[2])
		case "/help", "/start":
			mods.Help(botUrl, update.Message.Chat.ChatId)
		default:
			mods.SendMsg(botUrl, update.Message.Chat.ChatId, "Я не понимаю, воспользуйтесь /help")
		}

	} else {

		// Если пользователь отправил не сообщение:
		mods.SendMsg(botUrl, update.Message.Chat.ChatId, "Пока я воспринимаю только текст")

	}

}

// Функция инициализации конфига (всех токенов)
func initConfig() error {

	// Где конфиг
	viper.AddConfigPath("configs")

	// Как называется файл
	viper.SetConfigName("config")

	// Вывод статуса считывания (всё хорошо - вернёт nil)
	return viper.ReadInConfig()
}
