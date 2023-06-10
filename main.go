package main

import (
	"github.com/hud0shnik/weather_bot/internal/handler"
	"github.com/hud0shnik/weather_bot/internal/telegram"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func main() {

	// Настройка логгера
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Инициализация конфига (токенов)
	err := initConfig()
	if err != nil {
		logrus.Fatalf("initConfig error: %s", err)
		return
	}

	// Url бота для отправки и приёма сообщений
	botUrl := "https://api.telegram.org/bot" + viper.GetString("token")
	offSet := 0

	// Цикл работы бота
	for {

		// Получение апдейтов
		updates, err := telegram.GetUpdates(botUrl, offSet)
		if err != nil {
			logrus.Fatalf("getUpdates error: %s", err)
			return
		}

		// Обработка апдейтов
		for _, update := range updates {
			handler.Respond(botUrl, update)
			offSet = update.UpdateId + 1
		}

		// Вывод в консоль для тестов
		// fmt.Println(updates)
	}
}

// Функция инициализации конфига (всех токенов)
func initConfig() error {

	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
