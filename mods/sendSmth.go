package mods

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// Структуры для отправки сообщений, стикеров и картинок

type SendMessage struct {
	ChatId    int    `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type SendSticker struct {
	ChatId     int    `json:"chat_id"`
	StickerUrl string `json:"sticker"`
}

type SendPhoto struct {
	ChatId    int    `json:"chat_id"`
	PhotoUrl  string `json:"photo"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

// Функция отправки сообщения
func SendMsg(botUrl string, chatId int, msg string) error {

	// Формирование сообщения
	buf, err := json.Marshal(SendMessage{
		ChatId:    chatId,
		Text:      msg,
		ParseMode: "HTML",
	})
	if err != nil {
		log.Printf("json.Marshal error: %s", err)
		return err
	}

	// Отправка сообщения
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Printf("sendMessage error: %s", err)
		return err
	}

	return nil
}

// Функция отправки стикера
func SendStck(botUrl string, chatId int, url string) error {

	// Формирование стикера
	buf, err := json.Marshal(SendSticker{
		ChatId:     chatId,
		StickerUrl: url,
	})
	if err != nil {
		log.Printf("json.Marshal error: %s", err)
		return err
	}

	// Отправка стикера
	_, err = http.Post(botUrl+"/sendSticker", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Printf("sendSticker error: %s", err)
		return err
	}

	return nil
}

// Функция отправки картинки
func SendPict(botUrl string, chatId int, photoUrl, caption string) error {

	// Формирование картинки
	buf, err := json.Marshal(SendPhoto{
		ChatId:    chatId,
		PhotoUrl:  photoUrl,
		Caption:   caption,
		ParseMode: "HTML",
	})
	if err != nil {
		log.Printf("json.Marshal error: %s", err)
		return err
	}

	// Отправка картинки
	_, err = http.Post(botUrl+"/sendPhoto", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Printf("sendPhoto error: %s", err)
		return err
	}

	return nil
}
