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

// Функция для отправки сообщений пользователю
func SendMsg(botUrl string, update Update, msg string) error {

	// Запись того, что и куда отправить
	botMessage := SendMessage{
		ChatId: update.Message.Chat.ChatId,
		Text:   msg,
	}

	// Запись сообщения в json
	buf, err := json.Marshal(botMessage)

	// Проверка на ошибку
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("Marshal json error: ", err)
		return err
	}

	// Отправка сообщения
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))

	// Проверка на ошибку
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("SendMessage method error: ", err)
		return err
	}

	return nil
}

// Функция вывода информации о рассвете и закате
func Sun(botUrl string, update Update) error {

	// Получение координат из json'a
	lat, lon := getCoordinates(update)

	// Проверка на ошибку
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
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
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Чтение ответа
	body, _ := ioutil.ReadAll(res.Body)
	// Структура для записи ответа
	var rs = new(WeatherAPIResponse)
	// Запись ответа
	json.Unmarshal(body, &rs)

	// Вывод полученных данных пользователю
	SendMsg(botUrl, update, "🌄 Восход и закат на сегодня 🌄\n \n"+
		"🌅 Восход наступит в "+time.Unix(int64(rs.Current.Sunrise), 0).Add(3*time.Hour).Format("15:04:05")+
		"\n🌇 А закат в "+time.Unix(int64(rs.Current.Sunset), 0).Add(3*time.Hour).Format("15:04:05"))

	return nil
}

// Функция отправки почасовых карточек
func SendHourlyWeather(botUrl string, update Update, hours int) error {

	// Получение координат из json'a
	lat, lon := getCoordinates(update)

	// Проверка на ошибку
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// Реквест к openweathermap

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
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Чтение ответа
	body, _ := ioutil.ReadAll(res.Body)
	// Структура для записи ответа
	var rs1 = new(WeatherAPIResponse)
	// Запись ответа
	json.Unmarshal(body, &rs1)

	// Реквест к open-meteo

	// Ссылка к апи погоды
	url = "https://api.open-meteo.com/v1/forecast?latitude=" + lat + "&longitude=" + lon + "&hourly=temperature_2m,relativehumidity_2m,apparent_temperature,windspeed_10m&windspeed_unit=ms"
	// Генерация запроса
	req, _ = http.NewRequest("GET", url, nil)
	// Выполнение запроса
	res, err = http.DefaultClient.Do(req)

	// Проверка на ошибку
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Чтение ответа
	body, _ = ioutil.ReadAll(res.Body)
	// Структура для записи ответа
	var rs2 = new(openMeteoResponse)
	// Запись ответа
	json.Unmarshal(body, &rs2)

	// Вычисление средних значений и вывод полученных данных
	for n := 1; n < hours+1; n++ {
		SendMsg(botUrl, update, "Погода на "+time.Unix(rs1.Hourly[n].Dt, 0).Format("15:04")+":\n \n"+
			"На улице "+rs1.Hourly[n].Weather[0].Description+
			"\n🌡Температура: "+strconv.Itoa(int((rs1.Hourly[n].Temp+rs2.Hourly.Temperature[n])/2))+"°"+
			"\n🤔Ощущается как: "+strconv.Itoa(int((rs1.Hourly[n].Feels_like+rs2.Hourly.Feels_like[n])/2))+"°"+
			"\n💨Ветер: "+strconv.Itoa(int((rs1.Hourly[n].Wind_speed+rs2.Hourly.Wind_speed[n])/2))+" м/с"+
			"\n💧Влажность воздуха: "+strconv.Itoa((rs1.Hourly[n].Humidity+rs2.Hourly.Humidity[n])/2)+"%")
	}

	return nil
}

// Функция отправки дневных карточек
func SendDailyWeather(botUrl string, update Update, days int) error {

	// Получение координат из json'a
	lat, lon := getCoordinates(update)

	// Проверка на ошибку
	if lat == "err" {
		// Вывод и возврат ошибки
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
		return errors.New("wrong coordinates")
	}

	// Ссылка к апи погоды
	url := "https://api.openweathermap.org/data/2.5/onecall?lat=" + lat + "&lon=" + lon + "&lang=ru&exclude=minutely,current,minutely,alerts&units=metric&appid=" + viper.GetString("weatherToken")
	// Генерация запроса
	req, _ := http.NewRequest("GET", url, nil)
	// Выполнение запроса
	res, err := http.DefaultClient.Do(req)

	// Проверка на ошибку
	if err != nil {
		// Вывод и возврат ошибки
		fmt.Println("weather API error")
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Чтение ответа
	body, _ := ioutil.ReadAll(res.Body)
	// Структура для записи ответа
	var rs = new(WeatherAPIResponse)
	// Запись ответа
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

// Функция отправки конкретного прогноза и на два дня вперёд
func SendThreeDaysWeather(botUrl string, update Update) {

	// Если просто добавить в switch две команды,
	// то при некорректных данных будут выводиться две ошибки
	// Поэтому существует эта функция

	// Отправка текущего прогноза
	if SendCurrentWeather(botUrl, update) == nil {

		// Если всё хорошо, отправка двух дневных карточек
		SendDailyWeather(botUrl, update, 2)
	}
}

// Функция отправки погоды на данный момент
func SendCurrentWeather(botUrl string, update Update) error {

	// Получение координат из json'a
	lat, lon := getCoordinates(update)

	// Проверка на ошибку
	if lat == "err" {
		SendMsg(botUrl, update, "Пожалуйста обновите свои координаты командой /set")
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
		SendMsg(botUrl, update, "weather API error")
		return err
	}
	defer res.Body.Close()

	// Чтение ответа
	body, _ := ioutil.ReadAll(res.Body)
	// Структура для записи ответа
	var rs = new(WeatherAPIResponse)
	// Запись ответа
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

// Функция вывода списка команд
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

// Функция установки координат
func SetPlace(botUrl string, update Update) {

	// Открытие json файла для чтения координат
	file, err := os.Open("weather/coordinates.json")

	// Проверка на ошибку
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
	m[strconv.Itoa(update.Message.Chat.ChatId)] = update.Message.Text[5:]

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
	SendMsg(botUrl, update, "Записал координаты!")
}

// Функция получения координат
func getCoordinates(update Update) (string, string) {

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
	coords, c := m[strconv.Itoa(update.Message.Chat.ChatId)], 0

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

// Функция инициализации конфига (всех токенов)
func InitConfig() error {

	// Где конфиг
	viper.AddConfigPath("configs")

	// Как называется файл
	viper.SetConfigName("config")

	// Вывод статуса считывания (всё хорошо - вернёт nil)
	return viper.ReadInConfig()
}
