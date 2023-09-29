package app

import (
	"awesomeProject3/internal/config"
	"awesomeProject3/internal/handlers"
	"awesomeProject3/internal/nats"
	"awesomeProject3/internal/storage"
	"awesomeProject3/internal/ucache"
	"awesomeProject3/pkg/logging"
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter" // Импорт пакета для создания маршрутов HTTP.
	"net"                                 // Импорт пакета для работы с сетевыми соединениями.
	"net/http"                            // Импорт пакета для создания HTTP-сервера.
	"os"                                  // Импорт пакета для работы с операционной системой.
	"path"                                // Импорт пакета для работы с путями к файлам и директориям.
	"path/filepath"                       // Импорт пакета для работы с файловыми путями.
	"time"                                // Импорт пакета для работы с временем.
)

func Start() {
	// Получаем логгер для логирования информации о работе приложения.
	log := logging.GetLogger()

	// Инициализация компонентов приложения.

	log.Println("Init storage")
	storage := storage.NewStorage(log) // Инициализация хранилища (базы данных).
	conn := *storage.ConnectToDB()     // Подключение к базе данных.

	log.Println("Init cache")
	uCache := ucache.NewCache(log, storage) // Инициализация кэша.
	mcache := uCache.CreateCache(&conn)     // Создание кэша.

	log.Println("Cache created")

	log.Println("Init nats")
	nats := nats.NewNatsStreamingService(log, storage) // Инициализация службы Nats.
	sc := *nats.ConnectToNatsStreaming(&conn, mcache)  // Подключение к Nats Streaming.

	defer sc.Close()
	defer conn.Close(context.Background())

	// Создание маршрутизатора HTTP с использованием пакета httprouter.
	router := httprouter.New()

	// Получение конфигурации сервера.
	cfg := config.GetConfig()

	log.Info("Register server handler")

	// Создание обработчика HTTP-запросов и его регистрация для обработки маршрутов.
	handler := handlers.NewHandler(log, mcache)
	handler.Register(router)

	// Запуск веб-сервера с настройками из конфигурации.
	startRouter(router, cfg)
}

func startRouter(router *httprouter.Router, cfg *config.Config) {
	// Получение логгера для логирования информации о работе веб-сервера.
	logger := logging.GetLogger()
	logger.Info("Start application")

	var listener net.Listener
	var listenErr error

	// Если тип прослушивания задан как "sock" (Unix-сокет), то создаем и используем Unix-сокет.
	if cfg.Listen.Type == "sock" {
		logger.Info("detect app path")
		// Получение абсолютного пути к директории приложения.
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("create socket")
		// Формирование пути к Unix-сокету внутри директории приложения.
		socketPath := path.Join(appDir, "app.sock")
		logger.Debugf("socket path %s", socketPath)

		logger.Info("listen unix socket")
		// Прослушивание Unix-сокета.
		listener, err = net.Listen("unix", socketPath)
		logger.Infof("Server is listening unix socket: %s", socketPath)
	} else {
		// В противном случае, прослушивание TCP-порта, указанного в конфигурации.
		logger.Info("listen tcp")
		listener, listenErr = net.Listen("tcp", fmt.Sprintf("%s: %s", cfg.Listen.BindIp, cfg.Listen.Port))
		logger.Infof("Server is listening port %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)
	}

	if listenErr != nil {
		logger.Fatal(listenErr)
	}

	// Создание HTTP-сервера с настройками и маршрутизатором.
	server := &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Запуск сервера и обработка ошибки, если что-то идет не так.
	logger.Fatal(server.Serve(listener))
}
