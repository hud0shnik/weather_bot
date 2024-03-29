package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hud0shnik/weather_bot/internal/repository"
	"github.com/hud0shnik/weather_bot/internal/telegram"
	"github.com/sirupsen/logrus"
)

// Структуры для работы с Openweather API

type weatherAPIResponse struct {
	Current current `json:"current"`
	Daily   []day   `json:"daily"`
	Hourly  []hour  `json:"hourly"`
}

type current struct {
	Sunrise    int           `json:"sunrise"`
	Sunset     int           `json:"sunset"`
	Temp       float32       `json:"temp"`
	Feels_like float32       `json:"feels_like"`
	Humidity   int           `json:"humidity"`
	Wind_speed float32       `json:"wind_speed"`
	Weather    []weatherInfo `json:"weather"`
}

type day struct {
	Dt         int64         `json:"dt"`
	Sunrise    int           `json:"sunrise"`
	Sunset     int           `json:"sunset"`
	Temp       temp          `json:"temp"`
	Feels_like temp          `json:"feels_like"`
	Wind_speed float32       `json:"wind_speed"`
	Weather    []weatherInfo `json:"weather"`
	Humidity   int           `json:"humidity"`
}

type hour struct {
	Dt         int64         `json:"dt"`
	Temp       float32       `json:"temp"`
	Feels_like float32       `json:"feels_like"`
	Humidity   int           `json:"humidity"`
	Wind_speed float32       `json:"wind_speed"`
	Weather    []weatherInfo `json:"weather"`
}

type temp struct {
	Day     float32 `json:"day"`
	Night   float32 `json:"night"`
	Evening float32 `json:"eve"`
	Morning float32 `json:"morn"`
}

type weatherInfo struct {
	Description string `json:"description"`
}

// Функция вывода прогноза на три дня
func SendClassicWeather(botUrl string, chatId int) error {

	// Отправление прогноза на нынешнее время
	err := SendCurrentWeather(botUrl, chatId)
	if err != nil {
		return err
	}

	// Отправление прогноза на два следующих дня
	err = SendDailyWeather(botUrl, chatId, 2)
	if err != nil {
		return err
	}

	return nil
}

// Функция вывода информации о рассвете и закате
func SendSunInfo(botUrl string, chatId int) error {

	// Получение координат из json'a
	lat, lon, err := repository.GetCoordinates(chatId)
	if err != nil {
		telegram.SendMsg(botUrl, chatId, "Пожалуйста запишите свои координаты командой <b>/set</b>")
		return err
	}

	// Ссылка к апи погоды
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + os.Getenv("WEATHER_TOKEN"))
	if err != nil {
		logrus.Printf("http.Get error: %s", err)
		telegram.SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return err
	}
	defer resp.Body.Close()

	// Проверка респонса
	if resp.StatusCode != 200 {
		telegram.SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return err
	}

	// Запись респонса
	body, _ := io.ReadAll(resp.Body)
	var rs = new(weatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных пользователю
	telegram.SendMsg(botUrl, chatId, "🌄 Восход и закат на сегодня 🌄"+
		"\n🌅 Восход наступит в <i>"+time.Unix(int64(rs.Current.Sunrise), 0).Add(3*time.Hour).Format("15:04:05")+"</i>"+
		"\n🌇 А закат в <i>"+time.Unix(int64(rs.Current.Sunset), 0).Add(3*time.Hour).Format("15:04:05")+"</i>")

	return nil
}

// Функция отправки дневных карточек
func SendDailyWeather(botUrl string, chatId int, days int) error {

	// Получение координат из json'a
	lat, lon, err := repository.GetCoordinates(chatId)
	if err != nil {
		telegram.SendMsg(botUrl, chatId, "Пожалуйста запишите свои координаты командой <b>/set</b>")
		return err
	}

	// Отправка запроса API
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,current,minutely,alerts&units=metric&appid=" + os.Getenv("WEATHER_TOKEN"))
	if err != nil {
		logrus.Printf("http.Get error: %s", err)
		telegram.SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return err
	}
	defer resp.Body.Close()

	// Проверка респонса
	if resp.StatusCode != 200 {
		telegram.SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return err
	}

	// Запись респонса
	body, _ := io.ReadAll(resp.Body)
	var rs = new(weatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	for n := 1; n < days+1; n++ {
		telegram.SendMsg(botUrl, chatId, "Погода на <b>"+time.Unix(rs.Daily[n].Dt, 0).Format("02/01/2006")+"</b>:"+
			"\n----------------------------------------------"+
			"\n🌡Температура: <b>"+strconv.Itoa(int(rs.Daily[n].Temp.Morning))+"°</b>"+" -> <b>"+strconv.Itoa(int(rs.Daily[n].Temp.Evening))+"°</b>"+
			"\n🤔Ощущается как: <b>"+strconv.Itoa(int(rs.Daily[n].Feels_like.Morning))+"°</b>"+" -> <b>"+strconv.Itoa(int(rs.Daily[n].Feels_like.Evening))+"°</b>"+
			"\n💨Ветер: <b>"+strconv.Itoa(int(rs.Daily[n].Wind_speed))+" м/с</b>"+
			"\n💧Влажность воздуха: <b>"+strconv.Itoa(rs.Daily[n].Humidity)+"%</b>"+
			"\n----------------------------------------------")
	}

	return nil

}

// Функция отправки погоды на данный момент
func SendCurrentWeather(botUrl string, chatId int) error {

	// Получение координат из json'a
	lat, lon, err := repository.GetCoordinates(chatId)
	if err != nil {
		telegram.SendMsg(botUrl, chatId, "Пожалуйста запишите свои координаты командой <b>/set</b>")
		return err
	}

	// Ссылка к апи погоды
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + os.Getenv("WEATHER_TOKEN"))
	if err != nil {
		logrus.Printf("http.Get error: %s", err)
		telegram.SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return err
	}
	defer resp.Body.Close()

	// Проверка респонса
	if resp.StatusCode != 200 {
		telegram.SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return err
	}

	// Запись респонса
	body, _ := io.ReadAll(resp.Body)
	var rs = new(weatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	telegram.SendMsg(botUrl, chatId, "Погода <i>сейчас</i>"+":"+
		"\n----------------------------------------------"+
		"\n🌡Температура: <b>"+strconv.Itoa(int(rs.Current.Temp))+"</b>"+
		"\n🤔Ощущается как: <b>"+strconv.Itoa(int(rs.Current.Feels_like))+"°"+"</b>"+
		"\n💨Ветер: <b>"+strconv.Itoa(int(rs.Current.Wind_speed))+" м/с"+"</b>"+
		"\n💧Влажность воздуха: <b>"+strconv.Itoa(rs.Current.Humidity)+"%"+"</b>"+
		"\n----------------------------------------------")

	return nil
}
