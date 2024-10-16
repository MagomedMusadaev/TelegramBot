package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"testAPI/99_hw/logger"
	"time"
)

type ProductDB struct {
	db *sql.DB
}

var db *ProductDB
var infoLog = log.New(logger.LogFile, "INFO\t", log.Ldate|log.Ltime)
var errorLog = log.New(logger.LogFile, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

func ConnectProductDB() error {
	var err error
	dbConn, err := sql.Open("mysql", "root:7880000208mA+@tcp(127.0.0.1:8080)/testapi")
	if err != nil {
		return err
	}
	infoLog.Println("База данных запущена!")
	db = &ProductDB{db: dbConn} // Инициализация глобальной переменной
	return nil
}

func DB() *sql.DB {
	return db.db
}

func DailyCleanup() {
	for {
		now := time.Now()
		infoLog.Println("Проверка на полночь...")
		if now.Hour() == 0 && now.Minute() < 59 {
			infoLog.Println("Запущен процесс удаления данных с DB!")
			clearDatabase()
			time.Sleep(60 * time.Second)
		}
		time.Sleep(1 * time.Hour)
	}
}

func clearDatabase() {
	query := `DELETE FROM cryptocurrency_rates WHERE timestamp < (NOW() - INTERVAL 1 HOUR)` // сделать чтоб при очистке за последний час не удалял (DELETE)

	infoLog.Println("Удаление данных с DB...")

	_, err := db.db.Exec(query)
	if err != nil {
		errorLog.Println("Не удалось очистить базу данных!", err)
	} else {
		infoLog.Println("База данных успешно очищена!")
	}
}
