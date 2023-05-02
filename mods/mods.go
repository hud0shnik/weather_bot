package mods

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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
func Sun(botUrl string, chatId int) error {

	// Получение координат из json'a
	lat, lon := getCoordinates(chatId)

	// Проверка на ошибку
	if lat == "err" {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// Ссылка к апи погоды
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	// Генерация запроса
	req, _ := http.NewRequest("GET", url, nil)
	// Выполнение запроса
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("weather API error")
		SendMsg(botUrl, chatId, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Чтение ответа
	body, _ := ioutil.ReadAll(res.Body)
	// Структура для записи ответа
	var rs = new(weatherAPIResponse)
	// Запись ответа
	json.Unmarshal(body, &rs)

	// Вывод полученных данных пользователю
	SendMsg(botUrl, chatId, "🌄 Восход и закат на сегодня 🌄\n \n"+
		"🌅 Восход наступит в "+time.Unix(int64(rs.Current.Sunrise), 0).Add(3*time.Hour).Format("15:04:05")+
		"\n🌇 А закат в "+time.Unix(int64(rs.Current.Sunset), 0).Add(3*time.Hour).Format("15:04:05"))

	return nil
}

// Функция отправки дневных карточек
func SendDailyWeather(botUrl string, chatId int, days int) {

	// Получение координат из json'a
	lat, lon := getCoordinates(chatId)
	if lat == "err" {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой /set")
		return
	}

	// Отправка запроса API
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,current,minutely,alerts&units=metric&appid=" + viper.GetString("weatherToken"))
	if err != nil {
		fmt.Println("weather API error")
		SendMsg(botUrl, chatId, "weather API error")
		return
	}
	defer resp.Body.Close()

	// Проверка респонса
	if resp.StatusCode != 200 {
		SendMsg(botUrl, chatId, "weather API error")
		return
	}

	// Запись респонса
	body, _ := ioutil.ReadAll(resp.Body)
	var rs = new(weatherAPIResponse)
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	for n := 1; n < days+1; n++ {
		SendMsg(botUrl, chatId, "Погода на "+time.Unix(rs.Daily[n].Dt, 0).Format("02/01/2006")+":\n \n"+
			"\n----------------------------------------------"+
			"\n🌡Температура: "+strconv.Itoa(int(rs.Daily[n].Temp.Morning))+"°"+" -> "+strconv.Itoa(int(rs.Daily[n].Temp.Evening))+"°"+
			"\n🤔Ощущается как: "+strconv.Itoa(int(rs.Daily[n].Feels_like.Morning))+"°"+" -> "+strconv.Itoa(int(rs.Daily[n].Feels_like.Evening))+"°"+
			"\n💨Ветер: "+strconv.Itoa(int(rs.Daily[n].Wind_speed))+" м/с"+
			"\n💧Влажность воздуха: "+strconv.Itoa(rs.Daily[n].Humidity)+"%"+
			"\n----------------------------------------------")
	}

}

// Функция отправки погоды на данный момент
func SendCurrentWeather(botUrl string, chatId int) error {

	// Получение координат из json'a
	lat, lon := getCoordinates(chatId)

	// Проверка на ошибку
	if lat == "err" {
		SendMsg(botUrl, chatId, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// Ссылка к апи погоды
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,hourly,daily,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	// Генерация запроса
	req, _ := http.NewRequest("GET", url, nil)
	// Выполнение запроса
	res, err := http.DefaultClient.Do(req)

	// Проверка на ошибку
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("weather API error")
		SendMsg(botUrl, chatId, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Чтение ответа
	body, _ := ioutil.ReadAll(res.Body)
	// Структура для записи ответа
	var rs = new(weatherAPIResponse)
	// Запись ответа
	json.Unmarshal(body, &rs)

	// Вывод полученных данных
	SendMsg(botUrl, chatId, "Погода на сегодня"+":\n \n"+
		"\n----------------------------------------------"+
		"\n🌡Температура: "+strconv.Itoa(int(rs.Current.Temp))+
		"\n🤔Ощущается как: "+strconv.Itoa(int(rs.Current.Feels_like))+"°"+
		"\n💨Ветер: "+strconv.Itoa(int(rs.Current.Wind_speed))+" м/с"+
		"\n💧Влажность воздуха: "+strconv.Itoa(rs.Current.Humidity)+"%"+
		"\n----------------------------------------------")

	return nil
}

// Функция отправки конкретного прогноза и на два дня вперёд
func SendThreeDaysWeather(botUrl string, chatId int) {

	// Если просто добавить в switch две команды,
	// то при некорректных данных будут выводиться две ошибки
	// Поэтому существует эта функция

	// Отправка текущего прогноза
	if SendCurrentWeather(botUrl, chatId) == nil {

		// Если всё хорошо, отправка двух дневных карточек
		SendDailyWeather(botUrl, chatId, 2)
	}
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
func SetPlace(botUrl string, chatId int, coordinates string) {

	// Проверка на параметр
	if coordinates == "" {
		SendMsg(botUrl, chatId, "Вы не написали координаты, воспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return
	}

	// Открытие json файла для чтения координат
	file, err := os.Open("weather/coordinates.json")
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Map, в которую будет произведена запись всех координат
	var m map[string]string
	// Считывание текста файла
	body, _ := ioutil.ReadAll(file)
	// Запись в структуру
	json.Unmarshal(body, &m)

	// Добавление или обновление введенной информации в map
	m[strconv.Itoa(chatId)] = coordinates

	// Запись обновленных данных в json
	fileU, err := os.Create("weather/coordinates.json")

	// Проверка на ошибку
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer fileU.Close()

	// Форматирование координат в json
	result, _ := json.Marshal(m)
	// Запись в файл
	fileU.Write(result)

	//Уведомление об успешной записи данных
	SendMsg(botUrl, chatId, "Записал координаты!")
}

// Функция получения координат
func getCoordinates(chatId int) (string, string) {

	// Чтение данных из json файла с координатами
	file, err := os.Open("weather/coordinates.json")

	// Проверка на ошибку
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Map нужна для эффективной работы с данными
	// Ключ - айди диалога; Значение - введенные координаты
	var m map[string]string
	// Считывание текста файла
	body, _ := ioutil.ReadAll(file)
	// Запись в структуру
	json.Unmarshal(body, &m)

	// Получение координат, которые пользователь ввел ранее
	coords, c := m[strconv.Itoa(chatId)], 0

	// с - переменная, отвечающая за расположение пробела
	for ; c < len(coords); c++ {
		// Поиск пробела
		if coords[c] == ' ' {
			// Пробел найден, выход из цикла
			break
		}
	}

	// Если пробел в самом начале или его нет - ошибка
	if c == 0 || c == len(coords) {
		// Возврат ошибки
		return "err", "err"
	}

	// Получение координат
	latFloat, err := strconv.ParseFloat(coords[:c], 64)
	// Проверка. Широта не может быть больше 90 или меньше -90
	if err != nil || !(latFloat > -90 && latFloat < 90) {
		// Вывод ошибки
		return "err", "err"
	}

	// Получение координат
	lonFloat, err := strconv.ParseFloat(coords[c+1:], 64)
	// Проверка. У долготы тоже есть рамки: от -180 до 180
	if err != nil || !(lonFloat > -180 && lonFloat < 180) {
		// Вывод ошибки
		return "err", "err"
	}

	// Возврат координат
	return coords[:c], coords[c+1:]
}
