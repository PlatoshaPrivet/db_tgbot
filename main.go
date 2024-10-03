package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var changeID string = ""
var globalTable string = ""

type journal struct {
	id         int
	user_text  string
	status     string
	open_date  string
	close_date string
}

type tableList struct {
	tableName string
}

func telegramBot() {
	// Подключение к базе данных
	db, err := sql.Open("sqlite3", "data/tgbase.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Портируем бота
	bot, err := tgbotapi.NewBotAPI("YOUR_BOT_TOKEN")
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	currentTime := time.Now()

	// Обработка сообщений
	for update := range updates {
		if update.Message != nil && update.Message.NewChatMembers != nil && (update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup") {
			fmt.Printf("Found new chat member")
			for _, member := range update.Message.NewChatMembers {
				if member.UserName == "oajdgopanfnbo_bot" { // Бот был добавлен в группу
					groupName := update.Message.Chat.Title
					req := `CREATE TABLE IF NOT EXISTS '` + groupName + `' (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					user_text TEXT,
					status TEXT,
					open_date TEXT,
					close_date TEXT
					);`
					_, err := db.Exec(req)
					if err != nil {
						panic(err)
					}
					fmt.Printf("Database created for group: %s ", groupName)
				}
			}
		}
		if update.Message != nil && update.Message.Chat.Type == "private" && (update.Message.From.UserName == "AlexanderChekmarev" || update.Message.From.UserName == "T0n1_K" || update.Message.From.UserName == "fanatkakishlakadddd" || update.Message.From.UserName == "timoshvili") {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Выбрать базу данных", "opendb"),
							tgbotapi.NewInlineKeyboardButtonData("Заявки", "applic"),
							tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
						),
					)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Здравствуйте! Этот бот предназначен для управления заявками")
					msg.ReplyMarkup = inlineKeyboard

					bot.Send(msg)
				}
			}

			//Обработка #Описание
			if strings.HasPrefix(update.Message.Text, "#Описание ") {
				mess := update.Message.Text
				parts := strings.Split(mess, " ")
				id := parts[1]
				changeID = id
				row, err := db.Query("SELECT * FROM '"+globalTable+"' WHERE id = $1", id)
				if err != nil {
					panic(err)
				}

				for row.Next() {
					j := journal{}
					err := row.Scan(&j.id, &j.user_text, &j.status, &j.open_date, &j.close_date)
					if err != nil {
						panic(err)
					}
					var res = fmt.Sprintf("id: %d\nОписание: %s\nСтатус: %s\nДата открытия: %s  Дата закрытия: %s\n\n", j.id, j.user_text, j.status, j.open_date, j.close_date)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, res)
					bot.Send(msg)
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Действия:")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Изменить статус", "changestatus"),
						tgbotapi.NewInlineKeyboardButtonData("Удалить", "delete"),
						tgbotapi.NewInlineKeyboardButtonData("Вернуться к заявкам", "applic"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				bot.Send(msg)
			}
			//Обработка #СортДата
			if strings.HasPrefix(update.Message.Text, "#СортДата ") {
				mess := update.Message.Text
				parts := strings.Split(mess, " ")
				data1 := parts[1]
				data2 := parts[2]
				req := "SELECT * FROM '" + globalTable + "' WHERE open_date BETWEEN $1 AND $2"
				rows, err := db.Query(req, data1, data2)
				if err != nil {
					panic(err)
				}
				journals := []journal{}
				for rows.Next() {
					j := journal{}
					err := rows.Scan(&j.id, &j.user_text, &j.status, &j.open_date, &j.close_date)
					if err != nil {
						panic(err)
					}
					journals = append(journals, j)
				}
				var response []string
				for _, j := range journals {
					response = append(response, fmt.Sprintf("id: %d\nСтатус: %s\nДата открытия: %s  Дата закрытия: %s\n\n", j.id, j.status, j.open_date, j.close_date))
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Список отсортированных заявок по дате:\n"+strings.Join(response, "\n"))
				bot.Send(msg)
			}
			//Обрабока #СортДатаСтатус
			if strings.HasPrefix(update.Message.Text, "#СортДатаСтатус ") {
				mess := update.Message.Text
				parts := strings.Split(mess, " ")
				data1 := parts[1]
				data2 := parts[2]
				status := parts[3]
				fmt.Println(data1, " - ", data2, " - ", status)
				var req string
				if status == "Принято" {
					req = "SELECT * FROM '" + globalTable + "' WHERE status = 'Принято' AND open_date BETWEEN $1 AND $2"
				} else if status == "Завершено" {
					req = "SELECT * FROM '" + globalTable + "' WHERE open_date BETWEEN $1 AND $2 AND close_date BETWEEN $1 AND $2 AND status = 'Завершено'"
				}
				rows, err := db.Query(req, data1, data2)
				if err != nil {
					panic(err)
				}
				journals := []journal{}
				for rows.Next() {
					j := journal{}
					err := rows.Scan(&j.id, &j.user_text, &j.status, &j.open_date, &j.close_date)
					if err != nil {
						panic(err)
					}
					journals = append(journals, j)
				}
				var response []string
				for _, j := range journals {
					response = append(response, fmt.Sprintf("id: %d\nСтатус: %s\nДата открытия: %s  Дата закрытия: %s\n\n", j.id, j.status, j.open_date, j.close_date))
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Список отсортированных заявок по дате и статусу:\n"+strings.Join(response, "\n"))
				bot.Send(msg)
			}
		}

		if update.CallbackQuery != nil && update.CallbackQuery.Message.Chat.Type == "private" && (update.CallbackQuery.From.UserName == "AlexanderChekmarev" || update.CallbackQuery.From.UserName == "T0n1_K" || update.CallbackQuery.From.UserName == "fanatkakishlakadddd" || update.Message.From.UserName == "timoshvili") {
			log.Printf("Caught callback: %s", update.CallbackQuery.Data)

			rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
			if err != nil {
				log.Fatal(err)
			}
			for rows.Next() {
				t := tableList{}
				err := rows.Scan(&t.tableName)
				if err != nil {
					log.Fatal(err)
				}
				if update.CallbackQuery.Data == t.tableName {
					globalTable = t.tableName
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "База данных изменена")
					inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Заявки", "applic"),
							tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
						),
					)
					msg.ReplyMarkup = inlineKeyboard
					bot.Send(msg)
				}
			}

			//обработка "Заявки"
			if update.CallbackQuery.Data == "applic" && globalTable != "" {
				rows, err := db.Query("SELECT * FROM " + globalTable)
				if err != nil {
					panic(err)
				}
				journals := []journal{}
				for rows.Next() {
					j := journal{}
					err := rows.Scan(&j.id, &j.user_text, &j.status, &j.open_date, &j.close_date)
					if err != nil {
						panic(err)
					}
					journals = append(journals, j)
				}
				var response []string
				for _, j := range journals {
					response = append(response, fmt.Sprintf("id: %d\nСтатус: %s\nДата открытия: %s  Дата закрытия: %s\n\n", j.id, j.status, j.open_date, j.close_date))
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Список всех заявок:\n"+strings.Join(response, "\n"))
				msg2 := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Для просмотра описания заявки напишите: Описание [ID]\nДля сортировки по дате напишите: СортДата [в период с] [до] (формат даты: гггг-мм-дд)\nДля сортировки по дате и статусу напишите: СортДатаСтатус [в период с] [до] [статус заявки] (формат даты: гггг-мм-дд)")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("По статусу", "sortstatus"),
						tgbotapi.NewInlineKeyboardButtonData("Назад", "back"),
					),
				)
				msg2.ReplyMarkup = inlineKeyboard
				bot.Send(msg)
				bot.Send(msg2)
			} else if globalTable == "" {
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите базу данных!")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Назад", "back"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				bot.Send(msg)
			}
			//Обработка #СортСтатус
			if update.CallbackQuery.Data == "sortstatus" {
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Сортировать по статусу:")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Принято", "sortacc"),
						tgbotapi.NewInlineKeyboardButtonData("Завершено", "sortend"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				bot.Send(msg)

			}
			if update.CallbackQuery.Data == "sortacc" {
				rows, err := db.Query("SELECT * FROM '" + globalTable + "' WHERE status = 'Принято'")
				if err != nil {
					panic(err)
				}
				journals := []journal{}
				for rows.Next() {
					j := journal{}
					err := rows.Scan(&j.id, &j.user_text, &j.status, &j.open_date, &j.close_date)
					if err != nil {
						panic(err)
					}
					journals = append(journals, j)
				}
				var response []string
				for _, j := range journals {
					response = append(response, fmt.Sprintf("id: %d\nСтатус: %s\nДата открытия: %s  Дата закрытия: %s\n\n", j.id, j.status, j.open_date, j.close_date))
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Список отсортированных заявок по статусу:\n"+strings.Join(response, "\n"))
				bot.Send(msg)
			}
			if update.CallbackQuery.Data == "sortend" {
				rows, err := db.Query("SELECT * FROM '" + globalTable + "' WHERE status = 'Завершено'")
				if err != nil {
					panic(err)
				}
				journals := []journal{}
				for rows.Next() {
					j := journal{}
					err := rows.Scan(&j.id, &j.user_text, &j.status, &j.open_date, &j.close_date)
					if err != nil {
						panic(err)
					}
					journals = append(journals, j)
				}
				var response []string
				for _, j := range journals {
					response = append(response, fmt.Sprintf("id: %d\nСтатус: %s\nДата открытия: %s  Дата закрытия: %s\n\n", j.id, j.status, j.open_date, j.close_date))
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Список отсортированных заявок по статусу:\n"+strings.Join(response, "\n"))
				bot.Send(msg)
			}

			//Обработка "Помощь"
			if update.CallbackQuery.Data == "help" {
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Список команд, которые вы можете использовать:\n\n"+
					"Для общего чата:\n"+
					"1. [ответ на сообщение клиента] #Принято - Используется для изменения заявки на [Принято].\n"+
					"2. [ответ на сообщение клиента] #Решено - Используется для изменения заявки на [Завершено].\n\n"+
					"Для личного чата:\n"+
					"3.  - /start Перейти на главную страницу.\n"+
					"4.  - #СортДата позволяет отсортировать заявки по диапазону дат: #СортДата [в период с] [до] (формат даты: гггг-мм-дд).\n"+
					"5.  - #СортДатастатус позволяет отсортировать заявки по диапазону дат + статусу: #СортДатаСтатус [в период с] [до] [Принято/Завершено](формат даты: гггг-мм-дд)")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Назад", "back"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				bot.Send(msg)
			}
			//Обработка "Изменить статус" и дополнительные кнопки
			if update.CallbackQuery.Data == "changestatus" {
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите статус:")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Принято", "accept"),
						tgbotapi.NewInlineKeyboardButtonData("Завершено", "ended"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				bot.Send(msg)
			}
			if update.CallbackQuery.Data == "accept" {
				_, err := db.Exec(("update '" + globalTable + "' set status = 'Принято', close_date = '-' where id = $1"), changeID)
				if err != nil {
					panic(err)
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Статус изменен на 'Принято'")
				bot.Send(msg)
			}
			if update.CallbackQuery.Data == "ended" {
				req := "update '" + globalTable + "' set status = 'Завершено', close_date = $1 where id = $2"
				fmt.Print(req)
				_, err := db.Exec(req, currentTime.Format("2006-01-02"), changeID)
				if err != nil {
					panic(err)
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Статус изменен на 'Завершено'")
				bot.Send(msg)
			}
			//Обработка "Удалить"
			if update.CallbackQuery.Data == "delete" {
				_, err := db.Exec("delete from '"+globalTable+"' where id=$1", changeID)
				if err != nil {
					panic(err)
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Заявка удалена")
				bot.Send(msg)

			}

			//Обработка "Открыть базу данных"
			if update.CallbackQuery.Data == "opendb" {
				rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
				if err != nil {
					log.Fatal(err)
				}

				var buttons []tgbotapi.InlineKeyboardButton
				for rows.Next() {
					t := tableList{}
					err := rows.Scan(&t.tableName)
					if err != nil {
						log.Fatal(err)
					}
					if t.tableName == "sqlite_sequence" {
						continue
					}
					button := tgbotapi.NewInlineKeyboardButtonData(t.tableName, t.tableName)
					buttons = append(buttons, button)
				}

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите бд")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(buttons...),
				)
				msg.ReplyMarkup = inlineKeyboard

				bot.Send(msg)
			}
			if update.CallbackQuery.Data == "back" {
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Выбрать базу данных", "opendb"),
						tgbotapi.NewInlineKeyboardButtonData("Заявки", "applic"),
						tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
					),
				)

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Здравствуйте! Этот бот предназначен для управления заявками")
				msg.ReplyMarkup = inlineKeyboard

				bot.Send(msg)
			}
		} else if update.Message != nil && update.Message.ReplyToMessage != nil && (update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup") {

			if update.Message.Chat.IsSuperGroup() {
				//Обработка #Принято
				if update.Message.Text == "#Принято" || update.Message.Text == "#принято" {
					update.Message.MessageID = update.Message.ReplyToMessage.MessageID
					groupName := update.Message.Chat.Title
					req := "insert into '" + groupName + "' (user_text, open_date, close_date, status) values ($1, $2, '-', 'Принято')"
					_, err := db.Exec(req, update.Message.ReplyToMessage.Text, currentTime.Format("2006-01-02"))
					if err != nil {
						panic(err)
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваша заявка принята!")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
				//Обработка #Завершено
				if update.Message.Text == "#Завершено" || update.Message.Text == "#завершено" {
					update.Message.MessageID = update.Message.ReplyToMessage.MessageID
					req := "update '" + update.Message.Chat.Title + "' set status = 'Завершено', close_date = $1 where user_text = $2"
					_, err := db.Exec(req, currentTime.Format("2006-01-02"), update.Message.ReplyToMessage.Text)
					if err != nil {
						panic(err)
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваша заявка выполнена!")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
			}
		}

	}
}

func main() {
	telegramBot()
}
