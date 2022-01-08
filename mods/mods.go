package mods

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type TelegramResponse struct {
	Result []Update `json:"result"`
}

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ChatId int `json:"id"`
}

type SendMessage struct {
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

func SendMsg(botUrl string, update Update, msg string) error {
	botMessage := SendMessage{
		ChatId: update.Message.Chat.ChatId,
		Text:   msg,
	}
	buf, err := json.Marshal(botMessage)
	if err != nil {
		fmt.Println("Marshal json error: ", err)
		return err
	}
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println("SendMessage method error: ", err)
		return err
	}
	return nil
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}

type WeatherAPIResponse struct {
	Current Current `json:"current"`
	Daily   []Day   `json:"daily"`
}

type Day struct {
	Dt         int64         `json:"dt"`
	Sunrise    int           `json:"sunrise"`
	Sunset     int           `json:"sunset"`
	Temp       Temp          `json:"temp"`
	Feels_like Temp          `json:"feels_like"`
	Wind_speed float32       `json:"wind_speed"`
	Weather    []WeatherInfo `json:"weather"`
	Humidity   int           `json:"humidity"`
}

type Temp struct {
	/*
		"night 0,1,2,3,4,5",
		"morning 6,7,8,9,10,11",
		"day 12,13,14,15,16,17",
		"evening 18,19,20,21,22,23"
	*/
	Day     float32 `json:"day"`
	Night   float32 `json:"night"`
	Evening float32 `json:"eve"`
	Morning float32 `json:"morn"`
}

type Current struct {
	Sunrise    int           `json:"sunrise"`
	Sunset     int           `json:"sunset"`
	Temp       float32       `json:"temp"`
	Feels_like float32       `json:"feels_like"`
	Humidity   int           `json:"humidity"`
	Wind_speed float32       `json:"wind_speed"`
	Weather    []WeatherInfo `json:"weather"`
}

type WeatherInfo struct {
	Description string `json:"description"`
}

func Sun(botUrl string, update Update) error {
	lat, lon := getCoordinates(update)
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return nil
	}

	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var rs = new(WeatherAPIResponse)
	json.Unmarshal(body, &rs)

	result := "🌄 Восход и закат на сегодня 🌄\n \n" +
		"🌅 Восход наступит в " + time.Unix(int64(rs.Current.Sunrise), 0).Add(3*time.Hour).Format("15:04:05") +
		"\n🌇 А закат в " + time.Unix(int64(rs.Current.Sunset), 0).Add(3*time.Hour).Format("15:04:05")

	SendMsg(botUrl, update, result)
	return nil

}

func SendDailyWeather(botUrl string, update Update, days int) error {
	lat, lon := getCoordinates(update)
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return nil
	}

	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var rs = new(WeatherAPIResponse)
	json.Unmarshal(body, &rs)

	for n := 1; n < days+1; n++ {
		result := "Погода на " + time.Unix(rs.Daily[n].Dt, 0).Format("02/01/2006") + ":\n \n" +
			"На улице " + rs.Daily[n].Weather[0].Description +
			"\n🌡Температура: " + strconv.Itoa(int(rs.Daily[n].Temp.Morning)) + "°" + " -> " + strconv.Itoa(int(rs.Daily[n].Temp.Evening)) + "°" +
			"\n🤔Ощущается как: " + strconv.Itoa(int(rs.Daily[n].Feels_like.Morning)) + "°" + " -> " + strconv.Itoa(int(rs.Daily[n].Feels_like.Evening)) + "°" +
			"\n💨Ветер: " + strconv.Itoa(int(rs.Daily[n].Wind_speed)) + " м/с" +
			"\n💧Влажность воздуха: " + strconv.Itoa(rs.Daily[n].Humidity) + "%"

		SendMsg(botUrl, update, result)
	}
	return nil
}

func SendCurrentWeather(botUrl string, update Update) error {
	lat, lon := getCoordinates(update)
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return nil
	}
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var rs = new(WeatherAPIResponse)
	json.Unmarshal(body, &rs)

	result := "Погода на сегодня" + ":\n \n" +
		"На улице " + rs.Current.Weather[0].Description +
		"\n🌡Температура: " + strconv.Itoa(int(rs.Current.Temp)) +
		"\n🤔Ощущается как: " + strconv.Itoa(int(rs.Current.Feels_like)) + "°" +
		"\n💨Ветер: " + strconv.Itoa(int(rs.Current.Wind_speed)) + " м/с" +
		"\n💧Влажность воздуха: " + strconv.Itoa(rs.Current.Humidity) + "%"

	SendMsg(botUrl, update, result)
	SendDailyWeather(botUrl, update, 2)
	return nil
}

func SetPlace(botUrl string, update Update) {
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	m := map[string]string{}

	body, _ := ioutil.ReadAll(file)
	json.Unmarshal(body, &m)
	m[strconv.Itoa(update.Message.Chat.ChatId)] = update.Message.Text[5:]

	fileU, err := os.Create("weather/coordinates.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer fileU.Close()

	result, _ := json.Marshal(m)
	fileU.Write(result)

	fmt.Println("coordinates.json Updated!")
	SendMsg(botUrl, update, "Записал координаты!")
}

func getCoordinates(update Update) (string, string) {
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	m := map[string]string{}
	body, _ := ioutil.ReadAll(file)
	json.Unmarshal(body, &m)

	if len(m[strconv.Itoa(update.Message.Chat.ChatId)]) < 5 {
		return "err", "err"
	}

	coords, c := m[strconv.Itoa(update.Message.Chat.ChatId)], 0
	if coords == "err" {
		return "err", "err"
	}

	for ; c < len(coords); c++ {
		if coords[c] == ' ' {
			break
		}
	}

	if c == len(coords) || c == 0 {
		return "err", "err"
	}

	latFloat, err := strconv.ParseFloat(coords[:c], 64)
	if err != nil || latFloat > 90 || latFloat < -90 {
		return "err", "err"
	}
	lonFloat, err := strconv.ParseFloat(coords[c+1:], 64)
	if err != nil || lonFloat > 180 || lonFloat < -180 {
		return "err", "err"
	}

	return coords[:c], coords[c+1:]
}
