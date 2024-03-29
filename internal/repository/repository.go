package repository

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/hud0shnik/weather_bot/internal/telegram"
	"github.com/sirupsen/logrus"
)

// Функция установки координат
func SetCoordinates(botUrl string, chatId int, lat, lon string) {

	// Проверка на параметры
	if lat == "" || lon == "" {
		telegram.SendMsg(botUrl, chatId, "Вы не задали координаты, воспользуйтесь шаблоном ниже:")
		telegram.SendMsg(botUrl, chatId, "/set 55.5692101 37.4588852")
		return
	}

	// Проверка координат
	latFloat, err := strconv.ParseFloat(lat, 64)
	if err != nil || !(latFloat >= -90 && latFloat <= 90) {
		telegram.SendMsg(botUrl, chatId, "Широта (первый параметр) может принимать значения в диапазоне от <b>-90 до 90</b>.\nВоспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return
	}
	lonFloat, err := strconv.ParseFloat(lon, 64)
	if err != nil || !(lonFloat >= -180 && lonFloat <= 180) {
		telegram.SendMsg(botUrl, chatId, "Долгота (второй параметр) может принимать значения в диапазоне от <b>-180 до 180</b>.\nВоспользуйтесь шаблоном ниже:\n\n/set 55.5692101 37.4588852")
		return
	}

	// Открытие json файла для чтения координат
	file, err := os.Open("internal/repository/coordinates.json")
	if err != nil {
		logrus.Fatalf("Unable to open file: %s", err)
		return
	}
	defer file.Close()

	// Запись данных в карту
	var m map[string]string
	body, _ := io.ReadAll(file)
	json.Unmarshal(body, &m)

	// Обновление записей в карте
	m[strconv.Itoa(chatId)] = lat + " " + lon

	// Открытие файла
	fileU, err := os.Create("internal/repository/coordinates.json")
	if err != nil {
		logrus.Fatalf("Unable to create file: %s", err)
		return
	}
	defer fileU.Close()

	// Запись обновленных данных в json
	result, _ := json.Marshal(m)
	fileU.Write(result)

	telegram.SendMsg(botUrl, chatId, "Записал координаты!")
}

// Функция получения координат
func GetCoordinates(chatId int) (string, string, error) {

	// Чтение данных из json файла с координатами
	file, err := os.Open("internal/repository/coordinates.json")
	if err != nil {
		logrus.Fatalf("Unable to open file: %s", err)
		return "", "", err
	}
	defer file.Close()

	// Запись данных в структуру
	var m map[string]string
	body, _ := io.ReadAll(file)
	json.Unmarshal(body, &m)

	// Поиск и проверка на наличие записи
	coordsString, ok := m[strconv.Itoa(chatId)]
	if !ok {
		return "", "", errors.New("coordinates not found")
	}

	// Получение координат
	coords := strings.Fields(coordsString)

	return coords[0], coords[1], nil
}
