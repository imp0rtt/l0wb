package nats

import (
	"awesomeProject3/internal/model"   // Импорт модели данных приложения.
	"awesomeProject3/internal/storage" // Импорт пакета для работы с базой данных.
	"awesomeProject3/pkg/logging"      // Импорт пакета для логирования.
	"encoding/json"                    // Импорт пакета для работы с JSON.
	"github.com/jackc/pgx/v4"          // Импорт драйвера PostgreSQL.
	"github.com/nats-io/stan.go"       // Импорт пакета для работы с NATS Streaming.
	"github.com/patrickmn/go-cache"    // Импорт пакета для работы с кэшем.
	"time"                             // Импорт пакета для работы с временем.
)

// NatsStreamingService представляет сервис для работы с NATS Streaming.
type NatsStreamingService struct {
	log     *logging.Logger  // Логгер для записи событий.
	storage *storage.Storage // Хранилище данных (база данных).
}

// NewNatsStreamingService создает новый экземпляр NatsStreamingService.
func NewNatsStreamingService(log *logging.Logger, storage *storage.Storage) *NatsStreamingService {
	return &NatsStreamingService{log: log, storage: storage}
}

// ConnectToNatsStreaming устанавливает соединение с NATS Streaming сервером и подписывается на канал "server" для обработки сообщений.
func (n *NatsStreamingService) ConnectToNatsStreaming(conn *pgx.Conn, cache *cache.Cache) *stan.Conn {
	// Устанавливаем соединение с сервером NATS Streaming.
	n.log.Println("Подключение к NATS Streaming...")
	sc, _ := stan.Connect("test-cluster", "orderServer", stan.NatsURL("nats://localhost:4444"))

	// Подписываемся на канал "server" для дальнейшей обработки данных.
	sc.Subscribe("server", func(m *stan.Msg) {
		var order model.OrderInfo
		err := json.Unmarshal(m.Data, &order)
		if err != nil {
			n.log.Printf("Получены некорректные JSON-данные: %v", err)
		} else {
			n.log.Println("Получены корректные JSON-данные")
			cache.Set(order.OrderUid, string(m.Data), time.Duration(0))
			n.storage.InsertData(conn, order)
			n.log.Println("Данные из канала NATS:")
			n.log.Println(order.OrderUid, string(m.Data))
		}
	})

	n.log.Println("Подписка успешно оформлена!")
	return &sc
}
