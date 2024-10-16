package botAPI

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
	"strconv"
	"strings"
	"sync"
	"testAPI/99_hw/database"
	"testAPI/99_hw/logger"
	"time"
)

var mu sync.Mutex
var autoUpdateActive = false
var bot *tg.BotAPI

func BotAPI() {
	// Получаем токен из переменной окружения
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		logger.Error("TELEGRAM_TOKEN env variable not set")
		logger.Warning("Stop process!")
		os.Exit(1)
	}

	// Создаем нового бота
	var err error
	bot, err = tg.NewBotAPI(token)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	} else {
		logger.Info(fmt.Sprintf("Авторизация на аккаунт %s", bot.Self.UserName))
	}

	// Устанавливаем настройки для получения обновлений
	u := tg.NewUpdate(0)
	u.Timeout = 60

	// Получаем обновление
	updates, err := bot.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		if update.Message == nil { // игнорируем не текстовые сообщения
			continue
		}

		switch {
		case update.Message.Text == "/start":
			start(update)
		case update.Message.Text == "/rates":
			rates(update)
		case update.Message.Text == "/stop-auto":
			stopAuto(update)
		case strings.HasPrefix(update.Message.Text, "/start-auto"):
			parts := strings.Fields(update.Message.Text) // Разбиваем текст на части

			if len(parts) > 1 {
				minutesCount := strings.TrimSpace(parts[1]) // Получаем второй элемент
				minutes, err := strconv.Atoi(minutesCount)
				if err == nil && minutes > 0 {
					startAuto(update, minutes)
				} else {
					msg := tg.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите корректное количество минут, например: /start-auto 1")
					bot.Send(msg)
				}
			} else {
				msg := tg.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите количество минут, например: /start-auto 1")
				bot.Send(msg)
			}
		default:
			if strings.HasPrefix(update.Message.Text, "/rates") {
				cryptocurrency := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/rates"))
				getRate(update, cryptocurrency)
			}
		}
	}
}

func start(update tg.Update) {
	// Обрабатываем команду /start
	msg := tg.NewMessage(update.Message.Chat.ID, "Вечер в хату! Я бот для получения курсов валют!")
	msg2 := tg.NewMessage(update.Message.Chat.ID, "А так Зайнаб Махаеву знаешь?")

	commands := fmt.Sprintf(
		"'/start' - для начала работы!\n" +
			"'/rates' - для вывода всех курсов криптовалют\n" +
			"'/rates {название крипты}' - для вывода курса определенной криптовалюты\n" +
			"'/start-auto {minutes_count}' - для получения информации о всех криптовалютах через определенный промежуток времени\n" +
			"'/stop-auto' - для остановки автоматических обновлений\n",
	)

	msg3 := tg.NewMessage(update.Message.Chat.ID, commands)

	bot.Send(msg)
	bot.Send(msg2)
	bot.Send(msg3)
}

func rates(update tg.Update) {

	query := `SELECT currency_name, current_rate FROM cryptocurrency_rates ORDER BY id DESC LIMIT 2;`

	res, err := database.DB().Query(query)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса:" + err.Error())
		msg := tg.NewMessage(update.Message.Chat.ID, "Что-то пошло не так уцы, база моросит!")
		bot.Send(msg)
		return
	}
	defer res.Close()

	var rates []string
	for res.Next() {
		var currencyName string
		var currentRate float64

		if err := res.Scan(&currencyName, &currentRate); err != nil {
			logger.Error("Ошибка при считывании результата:" + err.Error())
			continue
		}
		rates = append(rates, fmt.Sprintf("%s: $%.2f", currencyName, currentRate))
	}

	if len(rates) == 0 {
		msg := tg.NewMessage(update.Message.Chat.ID, "Нет доступных курсов валют.")
		bot.Send(msg)
		return
	}

	msg := tg.NewMessage(update.Message.Chat.ID, "Текущие курсы:\n"+strings.Join(rates, "\n"))
	bot.Send(msg)

	MaxMin(update, "")
	Percent(update, "")
	return
}

func getRate(update tg.Update, cryptocurrency string) {

	var trigger string

	if strings.ToLower(cryptocurrency) == "биткоин" || strings.ToLower(cryptocurrency) == "bitcoin" {
		cryptocurrency = "bitcoin"
		trigger = "bitcoin"
	}
	if strings.ToLower(cryptocurrency) == "эфириум" || strings.ToLower(cryptocurrency) == "ethereum" {
		cryptocurrency = "ethereum"
		trigger = "ethereum"
	}

	query := `SELECT currency_name, current_rate FROM cryptocurrency_rates WHERE currency_name = ? ORDER BY id DESC LIMIT 1;`

	res, err := database.DB().Query(query, cryptocurrency)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса:" + err.Error())
		msg := tg.NewMessage(update.Message.Chat.ID, "Что пошло не так уцы, база моросит!")
		bot.Send(msg)
		return
	}
	defer res.Close()

	var rates []string
	for res.Next() {
		var currencyName string
		var currentRate float64

		if err := res.Scan(&currencyName, &currentRate); err != nil {
			fmt.Println("Ошибка при считывании результата:", err)
			continue
		}
		rates = append(rates, fmt.Sprintf("%s: $%.2f", currencyName, currentRate))
	}

	if len(rates) == 0 {
		msg := tg.NewMessage(update.Message.Chat.ID, "Нет доступных курсов валют.")
		bot.Send(msg)
		return
	}

	msg := tg.NewMessage(update.Message.Chat.ID, "Текущий курс:\n"+strings.Join(rates, "\n"))
	bot.Send(msg)

	MaxMin(update, trigger)
	Percent(update, trigger)

	return
}

func startAuto(update tg.Update, minutes int) {
	mu.Lock()
	defer mu.Unlock()

	if autoUpdateActive {
		msg := tg.NewMessage(update.Message.Chat.ID, "Автоматические обновления уже запущены!")
		bot.Send(msg)
		return
	}

	autoUpdateActive = true

	msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Настройки успешно сохранены:\n Установленое время %v минут!", minutes))
	bot.Send(msg)

	go func() {
		for {
			mu.Lock()
			if !autoUpdateActive {
				mu.Unlock()
				break
			}
			mu.Unlock()

			time.Sleep(time.Duration(minutes) * time.Minute)
			rates(update)
		}
	}()
}

func stopAuto(update tg.Update) {
	mu.Lock()
	autoUpdateActive = false
	mu.Unlock()
	msg := tg.NewMessage(update.Message.Chat.ID, "Автоматические обновления остановлены.")
	bot.Send(msg)
}

func MaxMin(update tg.Update, trigger string) {

	var query string
	switch trigger {
	case "":
		query = `SELECT currency_name, MAX(current_rate), MIN(current_rate) FROM cryptocurrency_rates WHERE DAYOFMONTH(timestamp) =  DAYOFMONTH(NOW()) GROUP BY currency_name`
	case "bitcoin":
		query = `SELECT currency_name, MAX(current_rate), MIN(current_rate) FROM cryptocurrency_rates WHERE DAYOFMONTH(timestamp) =  DAYOFMONTH(NOW()) AND currency_name = 'bitcoin' GROUP BY currency_name`
	default:
		if trigger == "ethereum" {
			query = `SELECT currency_name, MAX(current_rate), MIN(current_rate) FROM cryptocurrency_rates WHERE DAYOFMONTH(timestamp) =  DAYOFMONTH(NOW()) AND currency_name = 'ethereum' GROUP BY currency_name`
		}
	}

	res, err := database.DB().Query(query)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса:" + err.Error())
		msg := tg.NewMessage(update.Message.Chat.ID, "Что пошло не так уцы, база моросит!")
		bot.Send(msg)
		return
	}
	defer res.Close()

	var rates2 []string
	for res.Next() {
		var currencyName string
		var max, min, percentDifference float64

		if err := res.Scan(&currencyName, &max, &min); err != nil {
			logger.Error("Ошибка при считывании результата:" + err.Error())
			continue
		}

		if min > 0 {
			percentDifference = ((max - min) / min) * 100
			rates2 = append(rates2, fmt.Sprintf("%s: Макс: $%.2f, Мин: $%.2f, Разница: $%.2f (%.2f%%)", currencyName, max, min, max-min, percentDifference))
		} else {
			rates2 = append(rates2, fmt.Sprintf("%s: Макс: $%.2f, Мин: $%.2f, Разница: $%.2f (недоступно)", currencyName, max, min, max-min))
		}
	}

	if len(rates2) == 0 {
		msg := tg.NewMessage(update.Message.Chat.ID, "Нет доступных курсов валют.")
		bot.Send(msg)
		return
	}

	msg := tg.NewMessage(update.Message.Chat.ID, "Максимальные и минимальные курсы:\n"+strings.Join(rates2, "\n"))
	bot.Send(msg)

	return
}

func Percent(update tg.Update, trigger string) {

	var query string
	switch trigger {
	case "":
		query = `SELECT currency_name, MAX(current_rate) , MIN(current_rate) FROM cryptocurrency_rates WHERE timestamp >= (NOW() - INTERVAL 1 HOUR) GROUP BY currency_name; `
	case "bitcoin":
		query = `SELECT currency_name, MAX(current_rate) , MIN(current_rate) FROM cryptocurrency_rates WHERE timestamp >= (NOW() - INTERVAL 1 HOUR) AND currency_name = 'bitcoin'  GROUP BY currency_name; `
	default:
		if trigger == "ethereum" {
			query = `SELECT currency_name, MAX(current_rate) , MIN(current_rate) FROM cryptocurrency_rates WHERE timestamp >= (NOW() - INTERVAL 1 HOUR) AND currency_name = 'ethereum' GROUP BY currency_name; `
		}
	}

	res, err := database.DB().Query(query)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса:" + err.Error())
		msg := tg.NewMessage(update.Message.Chat.ID, "Что-то пошло не так уцы, база моросит!")
		bot.Send(msg)
		return
	}
	defer res.Close()

	var rates2 []string
	for res.Next() {
		var currencyName string
		var max, min, percentDifference float64

		if err := res.Scan(&currencyName, &max, &min); err != nil {
			logger.Error("Ошибка при считывании результата:" + err.Error())
			continue
		}

		if min > 0 {
			percentDifference = ((max - min) / min) * 100

			rates2 = append(rates2, fmt.Sprintf("%s: $%.2f (%.2f%%)", currencyName, max-min, percentDifference))
		} else {
			rates2 = append(rates2, fmt.Sprint("(недоступно)"))
		}
	}

	if len(rates2) == 0 {
		msg := tg.NewMessage(update.Message.Chat.ID, "Нет доступных курсов валют.")
		bot.Send(msg)
		return
	}

	msg := tg.NewMessage(update.Message.Chat.ID, fmt.Sprint("Изменение в % за последний час:\n"+strings.Join(rates2, "\n")))
	bot.Send(msg)

}
