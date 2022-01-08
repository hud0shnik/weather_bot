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
	err := mods.InitConfig()
	if err != nil {
		log.Println("Config error: ", err)
		return
	}
	botUrl := "https://api.telegram.org/bot" + viper.GetString("token")
	offSet := 0
	for {
		updates, err := getUpdates(botUrl, offSet)
		if err != nil {
			log.Println("Something went wrong: ", err)
		}
		for _, update := range updates {
			respond(botUrl, update)
			offSet = update.UpdateId + 1
		}
		fmt.Println(updates)
	}
}

func getUpdates(botUrl string, offset int) ([]mods.Update, error) {
	resp, err := http.Get(botUrl + "/getUpdates?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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

//	https://core.telegram.org/bots/api#using-a-local-bot-api-server
func respond(botUrl string, update mods.Update) error {
	msg := update.Message.Text

	switch msg {
	case "/week":
		mods.SendDailyWeather(botUrl, update, 7)
		return nil
	case "/weather":
		mods.SendCurrentWeather(botUrl, update)
		return nil
	case "/sun":
		mods.Sun(botUrl, update)
		return nil
	case "/today":
		mods.SendCurrentWeather(botUrl, update)
		return nil
	case "/set":
		mods.SendMsg(botUrl, update, "Вы не написали координаты, воспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return nil
	case "/help", "/start":
		mods.SendMsg(botUrl, update, "Команды: \n"+
			"/set - установить координаты\n"+
			"/weather - погода на сегодня и два следующих дня\n"+
			"/today - погода на сегодня\n"+
			"/week - погода на следующие 7 дней\n"+
			"/sun - время восхода и заката на сегодня")
		return nil
	}

	if len(msg) > 5 && msg[:4] == "/set" {
		mods.SetPlace(botUrl, update)
		mods.SendMsg(botUrl, update, "Введённые координаты: "+msg[4:])
		return nil
	}

	mods.SendMsg(botUrl, update, "Я не понимаю, воспользуйтесь /help")
	return nil
}
