package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	infoLog    *log.Logger
	warningLog *log.Logger
	errorLog   *log.Logger
	LogFile    *os.File
)

func Logging() {
	// Открываем файл для записи логов
	logFile, err := os.OpenFile("bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Не удалось открыть файл для логов!", err)
	}

	// Создаем MultiWriter для записи в файл и в консоль
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Инициализируем логгеры для разных уровней
	infoLog = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLog = log.New(multiWriter, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Записываем начальное сообщение
	infoLog.Println("Бот запущен")
}

func Info(msg string) {
	if infoLog != nil {
		infoLog.Println(msg)
	} else {
		fmt.Println("INFO LOGGING NOT INITIALIZED")
	}
}

func Warning(msg string) {
	if warningLog != nil {
		warningLog.Println(msg)
	} else {
		fmt.Println("WARNING LOGGING NOT INITIALIZED")
	}
}

func Error(msg string) {
	if errorLog != nil {
		errorLog.Println(msg)
	} else {
		fmt.Println("ERROR LOGGING NOT INITIALIZED")
	}
}
