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
	if err != nil {
		log.Println("Config error: ", err)
		return
	}
	// Url бота для отправки и приёма сообщений
	botUrl := "https://api.telegram.org/bot" + viper.GetString("token")
	offSet := 0

	for {
		// Получение апдейтов
		updates, err := getUpdates(botUrl, offSet)
		if err != nil {
			log.Println("Something went wrong: ", err)
		}

		// Обработка апдейтов
		for _, update := range updates {
			respond(botUrl, update)
			offSet = update.UpdateId + 1
		}

		// Вывод в консоль для тестов
		fmt.Println(updates)
	}
}

func getUpdates(botUrl string, offset int) ([]mods.Update, error) {
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
	var restResponse mods.TelegramResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}

	return restResponse.Result, nil
}

//	Функция обработки сообщений
func respond(botUrl string, update mods.Update) error {
	// msg - текст сообщения пользователя
	msg := update.Message.Text

	// Обработчик комманд
	switch msg {
	case "/week":
		mods.SendDailyWeather(botUrl, update, 7)
		return nil
	case "/hourly":
		mods.SendHourlyWeather(botUrl, update, 6)
		return nil
	case "/hourly24":
		mods.SendHourlyWeather(botUrl, update, 24)
		return nil
	case "/weather":
		mods.SendThreeDaysWeather(botUrl, update)
		return nil
	case "/current":
		mods.SendCurrentWeather(botUrl, update)
		return nil
	case "/sun":
		mods.Sun(botUrl, update)
		return nil
	case "/set":
		mods.SendMsg(botUrl, update, "Вы не написали координаты, воспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return nil
	case "/help", "/start":
		mods.Help(botUrl, update)
		return nil
	}

	// Команды, которые нельзя поместить в switch
	if len(msg) > 5 && msg[:4] == "/set" {
		mods.SetPlace(botUrl, update)
		mods.SendMsg(botUrl, update, "Введённые координаты: "+msg[4:])
		return nil
	}

	// Дефолтный респонс
	mods.SendMsg(botUrl, update, "Я не понимаю, воспользуйтесь /help")
	return nil
}
