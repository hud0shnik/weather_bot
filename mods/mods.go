package mods

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Структуры для работы с Telegram API
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

// Структуры для работы с Openweather API
type WeatherAPIResponse struct {
	Current Current `json:"current"`
	Daily   []Day   `json:"daily"`
	Hourly  []Hour  `json:"hourly"`
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

type Hour struct {
	Dt         int64         `json:"dt"`
	Temp       float32       `json:"temp"`
	Feels_like float32       `json:"feels_like"`
	Humidity   int           `json:"humidity"`
	Wind_speed float32       `json:"wind_speed"`
	Weather    []WeatherInfo `json:"weather"`
}

type Temp struct {
	Day     float32 `json:"day"`
	Night   float32 `json:"night"`
	Evening float32 `json:"eve"`
	Morning float32 `json:"morn"`
}

type WeatherInfo struct {
	Description string `json:"description"`
}

// Функция для отправки сообщений пользователю
func SendMsg(botUrl string, update Update, msg string) error {
	// Запись того, что и куда отправить
	botMessage := SendMessage{
		ChatId: update.Message.Chat.ChatId,
		Text:   msg,
	}

	// Запись сообщения в json
	buf, err := json.Marshal(botMessage)
	if err != nil {
		fmt.Println("Marshal json error: ", err)
		return err
	}

	// Отправка сообщения
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println("SendMessage method error: ", err)
		return err
	}
	return nil
}

func Sun(botUrl string, update Update) error {
	// Получение координат из json'a
	lat, lon := getCoordinates(update)
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// API реквест
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Запись ответа от API
	body, _ := ioutil.ReadAll(res.Body)
	var rs = new(WeatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных пользователю
	SendMsg(botUrl, update, "🌄 Восход и закат на сегодня 🌄\n \n"+
		"🌅 Восход наступит в "+time.Unix(int64(rs.Current.Sunrise), 0).Add(3*time.Hour).Format("15:04:05")+
		"\n🌇 А закат в "+time.Unix(int64(rs.Current.Sunset), 0).Add(3*time.Hour).Format("15:04:05"))

	return nil
}

func SendHourlyWeather(botUrl string, update Update, hours int) error {
	// Получение координат из json'a
	lat, lon := getCoordinates(update)
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// API реквест
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,daily,current,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Запись ответа от API
	body, _ := ioutil.ReadAll(res.Body)
	var rs = new(WeatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	for n := 1; n < hours+1; n++ {
		SendMsg(botUrl, update, "Погода на "+time.Unix(rs.Hourly[n].Dt, 0).Format("15:04")+":\n \n"+
			"На улице "+rs.Hourly[n].Weather[0].Description+
			"\n🌡Температура: "+strconv.Itoa(int(rs.Hourly[n].Temp))+"°"+
			"\n🤔Ощущается как: "+strconv.Itoa(int(rs.Hourly[n].Feels_like))+"°"+
			"\n💨Ветер: "+strconv.Itoa(int(rs.Hourly[n].Wind_speed))+" м/с"+
			"\n💧Влажность воздуха: "+strconv.Itoa(rs.Hourly[n].Humidity)+"%")
	}

	return nil
}

func SendDailyWeather(botUrl string, update Update, days int) error {
	// Получение координат из json'a
	lat, lon := getCoordinates(update)
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// API реквест
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,current,minutely,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Запись ответа от API
	body, _ := ioutil.ReadAll(res.Body)
	var rs = new(WeatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	for n := 1; n < days+1; n++ {
		SendMsg(botUrl, update, "Погода на "+time.Unix(rs.Daily[n].Dt, 0).Format("02/01/2006")+":\n \n"+
			"На улице "+rs.Daily[n].Weather[0].Description+
			"\n🌡Температура: "+strconv.Itoa(int(rs.Daily[n].Temp.Morning))+"°"+" -> "+strconv.Itoa(int(rs.Daily[n].Temp.Evening))+"°"+
			"\n🤔Ощущается как: "+strconv.Itoa(int(rs.Daily[n].Feels_like.Morning))+"°"+" -> "+strconv.Itoa(int(rs.Daily[n].Feels_like.Evening))+"°"+
			"\n💨Ветер: "+strconv.Itoa(int(rs.Daily[n].Wind_speed))+" м/с"+
			"\n💧Влажность воздуха: "+strconv.Itoa(rs.Daily[n].Humidity)+"%")
	}

	return nil
}

func SendThreeDaysWeather(botUrl string, update Update) {
	// Если просто добавить в switch две команды,
	// то при некорректных данных будут выводиться две ошибки
	if SendCurrentWeather(botUrl, update) == nil {
		SendDailyWeather(botUrl, update, 2)
	}
}

func SendCurrentWeather(botUrl string, update Update) error {
	// Получение координат из json'a
	lat, lon := getCoordinates(update)
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// API реквест
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Запись ответа от API
	body, _ := ioutil.ReadAll(res.Body)
	var rs = new(WeatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	SendMsg(botUrl, update, "Погода на сегодня"+":\n \n"+
		"На улице "+rs.Current.Weather[0].Description+
		"\n🌡Температура: "+strconv.Itoa(int(rs.Current.Temp))+
		"\n🤔Ощущается как: "+strconv.Itoa(int(rs.Current.Feels_like))+"°"+
		"\n💨Ветер: "+strconv.Itoa(int(rs.Current.Wind_speed))+" м/с"+
		"\n💧Влажность воздуха: "+strconv.Itoa(rs.Current.Humidity)+"%")

	return nil
}

func Help(botUrl string, update Update) {
	SendMsg(botUrl, update, "Команды: \n"+
		"/set - установить координаты\n"+
		"/weather - погода на сегодня и два следующих дня\n"+
		"/current - погода прямо сейчас\n"+
		"/hourly - погода на следующие 6 часов\n"+
		"/hourly24 - погода на следующие 24 часа\n"+
		"/week - погода на следующие 7 дней\n"+
		"/sun - время восхода и заката на сегодня")
}

func SetPlace(botUrl string, update Update) {
	// Открытие json файла для чтения координат
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Map, в которую будет произведена запись всех координат
	var m map[string]string
	body, _ := ioutil.ReadAll(file)
	json.Unmarshal(body, &m)

	// Добавление или обновление введенной информации в map
	m[strconv.Itoa(update.Message.Chat.ChatId)] = update.Message.Text[5:]

	// Запись обновленных данных в json
	fileU, err := os.Create("weather/coordinates.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer fileU.Close()
	result, _ := json.Marshal(m)
	fileU.Write(result)

	//Уведомление об успешной записи данных
	SendMsg(botUrl, update, "Записал координаты!")
}

func getCoordinates(update Update) (string, string) {
	// Чтение данных из json файла с координатами
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Map нужна для эффективной работы с данными
	// Ключ - айди диалога; Значение - введенные координаты
	var m map[string]string
	body, _ := ioutil.ReadAll(file)
	json.Unmarshal(body, &m)

	// Получение координат, которые пользователь ввел ранее
	coords, c := m[strconv.Itoa(update.Message.Chat.ChatId)], 0

	// с - переменная, отвечающая за расположение пробела
	for ; c < len(coords); c++ {
		if coords[c] == ' ' {
			break
		}
	}

	// Если пробел в самом начале или его нет - ошибка
	if c == 0 || c == len(coords) {
		return "err", "err"
	}

	// Широта не может быть больше 90 или меньше -90
	latFloat, err := strconv.ParseFloat(coords[:c], 64)
	if err != nil || !(latFloat > -90 && latFloat < 90) {
		return "err", "err"
	}

	// У долготы тоже есть рамки: от -180 до 180
	lonFloat, err := strconv.ParseFloat(coords[c+1:], 64)
	if err != nil || !(lonFloat > -180 && lonFloat < 180) {
		return "err", "err"
	}

	return coords[:c], coords[c+1:]
}

// Функция инициализации конфига (всех токенов)
func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
