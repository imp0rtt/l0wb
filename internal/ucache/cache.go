package ucache

import (
	"awesomeProject3/internal/storage"
	"awesomeProject3/pkg/logging"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/patrickmn/go-cache"
)

type Cache struct {
	log     *logging.Logger
	storage *storage.Storage
}

func NewCache(log *logging.Logger, storage *storage.Storage) *Cache {
	return &Cache{log: log, storage: storage}
}

func (c Cache) CreateCache(conn *pgx.Conn) *cache.Cache {
	order_uid, _ := c.storage.GetOrderUID(conn)
	Cache := cache.New(-1, -1)
	for i := range order_uid {
		data, _ := c.storage.GetDataByUID(conn, order_uid[i])
		Cache.Set(order_uid[i], data, cache.NoExpiration)
	}

	if len(Cache.Items()) > 0 {
		fmt.Printf("Обнаружено записей в бд: %v\n", len(Cache.Items()))
		fmt.Printf("Загрузка базы данных в кэш\n")
	} else {
		fmt.Println("Записей в бд не обнаружено")
	}

	return Cache
}
