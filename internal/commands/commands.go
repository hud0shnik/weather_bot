package commands

import "tgBot/internal/send"

// Функция вывода списка команд
func Help(botUrl string, chatId int) {
	send.SendMsg(botUrl, chatId, "Команды: \n"+
		"/set - установить координаты\n"+
		"/weather - погода на сегодня и два следующих дня\n"+
		"/current - погода прямо сейчас\n"+
		"/week - погода на следующие 7 дней\n"+
		"/sun - время восхода и заката на сегодня")
}
