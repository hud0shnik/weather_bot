package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"tgBot/mods"

	"github.com/spf13/viper"
)

func main() {

	// Инициализация конфига (токенов)
	err := mods.InitConfig()

	// Проверка на считывание конфига
	if err != nil {

		// Вывод ошибки
		log.Println("Config error: ", err)

		// Конец работы
		return
	}

	// Url бота для отправки и приёма сообщений
	botUrl := "https://api.telegram.org/bot" + viper.GetString("token")
	offSet := 0

	// Цикл работы бота
	for {

		// Получение апдейтов
		updates, err := getUpdates(botUrl, offSet)

		// Проверка на ошибку
		if err != nil {

			// Вывод ошибки
			log.Println("Something went wrong: ", err)

			// Конец работы бота
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
		fmt.Println(updates)
	}
}

// Функция получения апдейтов(сообщений)
func getUpdates(botUrl string, offset int) ([]mods.Update, error) {

	// Rest запрос для получения апдейтов
	resp, err := http.Get(botUrl + "/getUpdates?offset=" + strconv.Itoa(offset))

	// Проверка на ошибку
	if err != nil {

		// Возвращение ошибки
		return nil, err

	}
	defer resp.Body.Close()

	// Запись и обработка полученных данных
	body, err := ioutil.ReadAll(resp.Body)

	// Проверка на ошибку
	if err != nil {

		// Возврат ошибки
		return nil, err

	}

	// Структура для записи респонса
	var restResponse mods.TelegramResponse

	// Запись респонса в структуру
	err = json.Unmarshal(body, &restResponse)

	// Проверка на ошибку
	if err != nil {

		// Возврат ошибки
		return nil, err

	}

	// Возврат апдейтов
	return restResponse.Result, nil
}

//	Функция обработки сообщений
func respond(botUrl string, update mods.Update) error {

	// msg - текст сообщения пользователя
	msg := update.Message.Text

	// Обработчик комманд
	switch msg {

	// Если сообщение = /week
	case "/week":
		// Отправка семи карточек (по карточке на день)
		mods.SendDailyWeather(botUrl, update, 7)
		return nil

	// Если сообщение = /hourly
	case "/hourly":
		// Отправка шести карточек (по карточке на час)
		mods.SendHourlyWeather(botUrl, update, 6)
		return nil

	// Если сообщение = /hourly24
	case "/hourly24":
		// Отправка 24 карточек (по карточке на час)
		mods.SendHourlyWeather(botUrl, update, 24)
		return nil

	// Если сообщение = /weather
	case "/weather":
		// Отправка трёх карточек (по карточке на день)
		mods.SendThreeDaysWeather(botUrl, update)
		return nil

	// Если сообщение = /current
	case "/current":
		// Отправка погоды на данный момент
		mods.SendCurrentWeather(botUrl, update)
		return nil

	// Если сообщение = /sun
	case "/sun":
		// Отправка времени рассвета и заката
		mods.Sun(botUrl, update)
		return nil

	// Если сообщение = /set (без параметра)
	case "/set":
		// Уведомление о том, как нужно пользоваться этой функцией
		mods.SendMsg(botUrl, update, "Вы не написали координаты, воспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return nil

	// Если сообщение = /help или /start
	case "/help", "/start":
		// Вывод списка команд
		mods.Help(botUrl, update)
		return nil
	}

	// Обработка команды /set с параметром
	if len(msg) > 5 && msg[:4] == "/set" {

		// Запись координат
		mods.SetPlace(botUrl, update)

		// Уведомление для пользователя
		mods.SendMsg(botUrl, update, "Введённые координаты: "+msg[4:])

		// Всё хорошо, возврат нулевой ошибки
		return nil

	}

	// Дефолтный респонс
	mods.SendMsg(botUrl, update, "Я не понимаю, воспользуйтесь /help")
	return nil
}
