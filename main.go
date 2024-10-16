package main

import (
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"testAPI/99_hw/botAPI"
	"testAPI/99_hw/database"
	"testAPI/99_hw/logger"
	"testAPI/99_hw/requestDecodUpdate"
	"time"
)

func main() {

	wg := sync.WaitGroup{}

	logger.Logging()
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum&vs_currencies=usd"

	interval := 5 * time.Minute

	if err := database.ConnectProductDB(); err != nil {
		logger.Error("Ошибка подключения к базе данных:" + err.Error())
		return
	}
	defer database.DB().Close()
	defer logger.LogFile.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		requestDecodUpdate.UpdatePrices(url, interval)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		botAPI.BotAPI()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		database.DailyCleanup()
	}()

	wg.Wait()
}
