package mods

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
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

// Структуры для работы с Open-Meteo API

type openMeteoResponse struct {
	Error  bool            `json:"error"`
	Hourly openMeteoHourly `json:"hourly"`
}

type openMeteoHourly struct {
	Temperature []float32 `json:"temperature_2m"`
	Humidity    []int     `json:"relativehumidity_2m"`
	Feels_like  []float32 `json:"apparent_temperature"`
	Wind_speed  []float32 `json:"windspeed_10m"`
}

// Функция вывода информации о рассвете и закате
func SendSunInfo(botUrl string, chatId int) {

	// Получение координат из json'a
	lat, lon, err := getCoordinates(chatId)
	if err != nil {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой /set")
		return
	}

	// Ссылка к апи погоды
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + viper.GetString("weatherToken"))
	if err != nil {
		log.Printf("http.Get error: %s", err)
		SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return
	}
	defer resp.Body.Close()

	// Проверка респонса
	if resp.StatusCode != 200 {
		SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return
	}

	// Запись респонса
	body, _ := ioutil.ReadAll(resp.Body)
	var rs = new(weatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных пользователю
	SendMsg(botUrl, chatId, "🌄 Восход и закат на сегодня 🌄"+
		"\n🌅 Восход наступит в "+time.Unix(int64(rs.Current.Sunrise), 0).Add(3*time.Hour).Format("15:04:05")+
		"\n🌇 А закат в "+time.Unix(int64(rs.Current.Sunset), 0).Add(3*time.Hour).Format("15:04:05"))

}

// Функция отправки дневных карточек
func SendDailyWeather(botUrl string, chatId int, days int) {

	// Получение координат из json'a
	lat, lon, err := getCoordinates(chatId)
	if err != nil {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой /set")
		return
	}

	// Отправка запроса API
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,current,minutely,alerts&units=metric&appid=" + viper.GetString("weatherToken"))
	if err != nil {
		log.Printf("http.Get error: %s", err)
		SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return
	}
	defer resp.Body.Close()

	// Проверка респонса
	if resp.StatusCode != 200 {
		SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return
	}

	// Запись респонса
	body, _ := ioutil.ReadAll(resp.Body)
	var rs = new(weatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	for n := 1; n < days+1; n++ {
		SendMsg(botUrl, chatId, "Погода на "+time.Unix(rs.Daily[n].Dt, 0).Format("02/01/2006")+":"+
			"\n----------------------------------------------"+
			"\n🌡Температура: "+strconv.Itoa(int(rs.Daily[n].Temp.Morning))+"°"+" -> "+strconv.Itoa(int(rs.Daily[n].Temp.Evening))+"°"+
			"\n🤔Ощущается как: "+strconv.Itoa(int(rs.Daily[n].Feels_like.Morning))+"°"+" -> "+strconv.Itoa(int(rs.Daily[n].Feels_like.Evening))+"°"+
			"\n💨Ветер: "+strconv.Itoa(int(rs.Daily[n].Wind_speed))+" м/с"+
			"\n💧Влажность воздуха: "+strconv.Itoa(rs.Daily[n].Humidity)+"%"+
			"\n----------------------------------------------")
	}

}

// Функция отправки погоды на данный момент
func SendCurrentWeather(botUrl string, chatId int) {

	// Получение координат из json'a
	lat, lon, err := getCoordinates(chatId)
	if err != nil {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой /set")
		return
	}

	// Ссылка к апи погоды
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + viper.GetString("weatherToken"))
	if err != nil {
		log.Printf("http.Get error: %s", err)
		SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return
	}
	defer resp.Body.Close()

	// Проверка респонса
	if resp.StatusCode != 200 {
		SendMsg(botUrl, chatId, "Внутренняя ошибка")
		return
	}

	// Запись респонса
	body, _ := ioutil.ReadAll(resp.Body)
	var rs = new(weatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	SendMsg(botUrl, chatId, "Погода сейчас"+":"+
		"\n----------------------------------------------"+
		"\n🌡Температура: "+strconv.Itoa(int(rs.Current.Temp))+
		"\n🤔Ощущается как: "+strconv.Itoa(int(rs.Current.Feels_like))+"°"+
		"\n💨Ветер: "+strconv.Itoa(int(rs.Current.Wind_speed))+" м/с"+
		"\n💧Влажность воздуха: "+strconv.Itoa(rs.Current.Humidity)+"%"+
		"\n----------------------------------------------")

}

// Функция вывода списка команд
func Help(botUrl string, chatId int) {
	SendMsg(botUrl, chatId, "Команды: \n"+
		"/set - установить координаты\n"+
		"/weather - погода на сегодня и два следующих дня\n"+
		"/current - погода прямо сейчас\n"+
		"/week - погода на следующие 7 дней\n"+
		"/sun - время восхода и заката на сегодня")
}

// Функция установки координат
func SetPlace(botUrl string, chatId int, lat, lon string) {

	// Проверка на параметр
	if lat == "" || lon == "" {
		SendMsg(botUrl, chatId, "Вы не написали координаты, воспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return
	}

	// Открытие json файла для чтения координат
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		log.Fatalf("Unable to open file: %v", err)
		return
	}
	defer file.Close()

	// Запись данных в карту
	var m map[string]string
	body, _ := ioutil.ReadAll(file)
	json.Unmarshal(body, &m)

	// Обновление введенной информации
	m[strconv.Itoa(chatId)] = lat + " " + lon

	// Открытие файла
	fileU, err := os.Create("weather/coordinates.json")
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
		return
	}
	defer fileU.Close()

	// Запись обновленных данных в json
	result, _ := json.Marshal(m)
	fileU.Write(result)

	SendMsg(botUrl, chatId, "Записал координаты!")
}

// Функция получения координат
func getCoordinates(chatId int) (string, string, error) {

	// Чтение данных из json файла с координатами
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		log.Fatalf("Unable to open file: %v", err)
		return "", "", err
	}
	defer file.Close()

	// Запись данных в структуру
	var m map[string]string
	body, _ := ioutil.ReadAll(file)
	json.Unmarshal(body, &m)

	// Получение координат
	coords := strings.Fields(m[strconv.Itoa(chatId)])

	// Проверка координат
	latFloat, err := strconv.ParseFloat(coords[0], 64)
	if err != nil || !(latFloat > -90 && latFloat < 90) {
		return "", "", err
	}
	lonFloat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil || !(lonFloat > -180 && lonFloat < 180) {
		return "", "", err
	}

	return coords[0], coords[1], nil
}
