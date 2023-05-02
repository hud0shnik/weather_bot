package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
	err := mods.InitConfig()
	if err != nil {
		log.Println("Config error: ", err)
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
			log.Println("Something went wrong: ", err)
			return
		}

		// Обработка апдейтов
		for _, update := range updates {

			// Вызов функции генерации ответа
			respond(botUrl, update)

			// Обновление счётчика апдейтов
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
func respond(botUrl string, update update) error {

	// msg - текст сообщения пользователя
	msg := update.Message.Text

	// Обработчик комманд
	switch msg {

	// Если сообщение = /week
	case "/week":
		// Отправка семи карточек (по карточке на день)
		mods.SendDailyWeather(botUrl, update.Message.Chat.ChatId, 7)
		return nil

	// Если сообщение = /weather
	case "/weather":
		// Отправка трёх карточек (по карточке на день)
		mods.SendThreeDaysWeather(botUrl, update.Message.Chat.ChatId)
		return nil

	// Если сообщение = /current
	case "/current":
		// Отправка погоды на данный момент
		mods.SendCurrentWeather(botUrl, update.Message.Chat.ChatId)
		return nil

	// Если сообщение = /sun
	case "/sun":
		// Отправка времени рассвета и заката
		mods.Sun(botUrl, update.Message.Chat.ChatId)
		return nil

	// Если сообщение = /set (без параметра)
	case "/set":
		// Уведомление о том, как нужно пользоваться этой функцией
		mods.SendMsg(botUrl, update.Message.Chat.ChatId, "Вы не написали координаты, воспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return nil

	// Если сообщение = /help или /start
	case "/help", "/start":
		// Вывод списка команд
		mods.Help(botUrl, update.Message.Chat.ChatId)
		return nil
	}

	// Обработка команды /set с параметром
	if len(msg) > 5 && msg[:4] == "/set" {

		// Запись координат
		mods.SetPlace(botUrl, update.Message.Chat.ChatId, update.Message.Text[5:])

		// Уведомление для пользователя
		mods.SendMsg(botUrl, update.Message.Chat.ChatId, "Введённые координаты: "+msg[4:])

		// Всё хорошо, возврат нулевой ошибки
		return nil

	}

	// Дефолтный респонс
	mods.SendMsg(botUrl, update.Message.Chat.ChatId, "Я не понимаю, воспользуйтесь /help")
	return nil
}
