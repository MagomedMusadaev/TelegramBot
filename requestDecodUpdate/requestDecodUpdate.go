package requestDecodUpdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testAPI/99_hw/database"
	"testAPI/99_hw/logger"
	"time"
)

func decoding(resp *http.Response) map[string]map[string]float64 {

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("Ошибка декодирования:" + err.Error())
	} else {
		logger.Info("Декодирование прошло успешно")
	}
	defer resp.Body.Close()
	return result
}

func request(url string) *http.Response {

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("URL запрос не был принят!" + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Ошибка при получении данных: %d", resp.StatusCode))
	}
	return resp
}

func UpdatePrices(url string, interval time.Duration) {
	for {
		resp := request(url)
		if resp != nil {
			result := decoding(resp)

			for currency, data := range result {

				if price, ok := data["usd"]; ok {

					query := `INSERT INTO cryptocurrency_rates (currency_name, current_rate, timestamp) VALUES (?, ?, ?)`

					tm := time.Now().Add(3 * time.Hour)

					_, err := database.DB().Exec(query, currency, price, tm)
					if err != nil {
						logger.Error(fmt.Sprintf("Ошибка при записи %s: %v\n", currency, err))
					} else {
						logger.Info(fmt.Sprintf("Данные для %s успешно сохранены: $%.2f\n", currency, price))
					}
				} else {
					logger.Warning(fmt.Sprintf("%s: данные недоступны\n", currency))
				}
			}
			resp.Body.Close() // Закрываем тело ответа
		}
		time.Sleep(interval) // Ждем 5 минут
	}
}
