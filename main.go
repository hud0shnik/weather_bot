package main

import (
	"log"
	"tgBot/internal/handler"
	"tgBot/internal/telegram"

	"github.com/spf13/viper"
)

func main() {

	// Инициализация конфига (токенов)
	err := initConfig()
	if err != nil {
		log.Fatalf("initConfig error: %s", err)
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
			log.Fatalf("getUpdates error: %s", err)
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
