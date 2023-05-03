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

// Функция вывода информации о рассвете и закате
func SendSunInfo(botUrl string, chatId int) {

	// Получение координат из json'a
	lat, lon, err := getCoordinates(chatId)
	if err != nil {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой <b>/set</b>")
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
		"\n🌅 Восход наступит в <i>"+time.Unix(int64(rs.Current.Sunrise), 0).Add(3*time.Hour).Format("15:04:05")+"</i>"+
		"\n🌇 А закат в <i>"+time.Unix(int64(rs.Current.Sunset), 0).Add(3*time.Hour).Format("15:04:05")+"</i>")

}

// Функция отправки дневных карточек
func SendDailyWeather(botUrl string, chatId int, days int) {

	// Получение координат из json'a
	lat, lon, err := getCoordinates(chatId)
	if err != nil {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой <b>/set</b>")
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
		SendMsg(botUrl, chatId, "Погода на <b>"+time.Unix(rs.Daily[n].Dt, 0).Format("02/01/2006")+"</b>:"+
			"\n----------------------------------------------"+
			"\n🌡Температура: <b>"+strconv.Itoa(int(rs.Daily[n].Temp.Morning))+"°</b>"+" -> <b>"+strconv.Itoa(int(rs.Daily[n].Temp.Evening))+"°</b>"+
			"\n🤔Ощущается как: <b>"+strconv.Itoa(int(rs.Daily[n].Feels_like.Morning))+"°</b>"+" -> <b>"+strconv.Itoa(int(rs.Daily[n].Feels_like.Evening))+"°</b>"+
			"\n💨Ветер: <b>"+strconv.Itoa(int(rs.Daily[n].Wind_speed))+" м/с</b>"+
			"\n💧Влажность воздуха: <b>"+strconv.Itoa(rs.Daily[n].Humidity)+"%</b>"+
			"\n----------------------------------------------")
	}

}

// Функция отправки погоды на данный момент
func SendCurrentWeather(botUrl string, chatId int) {

	// Получение координат из json'a
	lat, lon, err := getCoordinates(chatId)
	if err != nil {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой <b>/set</b>")
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
	SendMsg(botUrl, chatId, "Погода <i>сейчас</i>"+":"+
		"\n----------------------------------------------"+
		"\n🌡Температура: <b>"+strconv.Itoa(int(rs.Current.Temp))+"</b>"+
		"\n🤔Ощущается как: <b>"+strconv.Itoa(int(rs.Current.Feels_like))+"°"+"</b>"+
		"\n💨Ветер: <b>"+strconv.Itoa(int(rs.Current.Wind_speed))+" м/с"+"</b>"+
		"\n💧Влажность воздуха: <b>"+strconv.Itoa(rs.Current.Humidity)+"%"+"</b>"+
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

	// Проверка на параметры
	if lat == "" || lon == "" {
		SendMsg(botUrl, chatId, "Вы не написали координаты, воспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return
	}

	// Проверка координат
	latFloat, err := strconv.ParseFloat(lat, 64)
	if err != nil || !(latFloat > -90 && latFloat < 90) {
		SendMsg(botUrl, chatId, "Широта (первый параметр) может принимать значения в диапазоне от <b>-90 до 90</b>.\nВоспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return
	}
	lonFloat, err := strconv.ParseFloat(lon, 64)
	if err != nil || !(lonFloat > -180 && lonFloat < 180) {
		SendMsg(botUrl, chatId, "Долгота (второй параметр) может принимать значения в диапазоне от <b>-180 до 180</b>.\nВоспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return
	}

	// Открытие json файла для чтения координат
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		log.Fatalf("Unable to open file: %s", err)
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
		log.Fatalf("Unable to create file: %s", err)
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
		log.Fatalf("Unable to open file: %s", err)
		return "", "", err
	}
	defer file.Close()

	// Запись данных в структуру
	var m map[string]string
	body, _ := ioutil.ReadAll(file)
	json.Unmarshal(body, &m)

	// Получение координат
	coords := strings.Fields(m[strconv.Itoa(chatId)])

	return coords[0], coords[1], nil
}
