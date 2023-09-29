package storage

import (
	"awesomeProject3/internal/model"
	"awesomeProject3/pkg/logging"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"os"
)

// Storage представляет сервис для работы с базой данных PostgreSQL.
type Storage struct {
	log *logging.Logger // Логгер для записи событий.
}

// NewStorage создает новый экземпляр Storage.
func NewStorage(log *logging.Logger) *Storage {
	return &Storage{log: log}
}

// ConnectToDB устанавливает соединение с базой данных PostgreSQL.
func (s *Storage) ConnectToDB() *pgx.Conn {
	err := godotenv.Load()
	if err != nil {
		s.log.Println("Не найден файл .env")
		os.Exit(1)
	}
	// Формируем строку подключения к базе данных из переменных окружения.
	url := fmt.Sprintf("%v://%v:%v@%v:%v/%v", os.Getenv("DRIVER"), os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"), os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("DB_NAME"))
	// Устанавливаем соединение с базой данных.
	connect, err := pgx.Connect(context.Background(), url)
	if err != nil {
		s.log.Printf("Не удалось подключиться к базе данных: %v\n", err)
		os.Exit(1)
	}

	return connect
}

// GetOrderUID получает список уникальных идентификаторов заказов из базы данных.
func (s *Storage) GetOrderUID(conn *pgx.Conn) ([]string, error) {
	var UIDs []string
	query := `SELECT array_agg(order_uid) FROM order_info`

	err := conn.QueryRow(context.Background(), query).Scan(&UIDs)
	return UIDs, err
}

// GetDataByUID извлекает информацию о заказе по его уникальному идентификатору.
func (s *Storage) GetDataByUID(conn *pgx.Conn, order_uid string) (string, error) {
	var order model.OrderInfo
	query := `WITH DeliveryInfo AS (
    SELECT
        d.name,
        d.phone,
        d.zip,
        d.city,
        d.address,
        d.region,
        d.email
    FROM deliveries d
    WHERE d.id = (
        SELECT od.delivery_id
        FROM order_delivery od
        WHERE od.order_uid = $1
    )
),
ItemsInfo AS (
    SELECT jsonb_agg(to_jsonb(i.*)) AS "items"
    FROM items i
    WHERE i.track_number = (
        SELECT oi.track_number
        FROM order_info oi
        WHERE oi.order_uid = $1
    )
)
SELECT
    oi.*,
    to_jsonb(p.*) AS "payment",
    di."items" AS "items",
    to_jsonb(del) AS "delivery"
FROM order_info oi
LEFT JOIN payments p ON p."transaction" = oi.order_uid
JOIN DeliveryInfo del ON true
JOIN ItemsInfo di ON true
WHERE oi.order_uid = $1
LIMIT 1;
`
	if err := conn.QueryRow(context.Background(), query, order_uid).Scan(
		&order.OrderUid,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerId,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmId,
		&order.DateCreated,
		&order.OofShard,
		&order.Payment,
		&order.Items,
		&order.Delivery,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return "", pgErr
		}
	}
	res, _ := json.Marshal(&order)

	return string(res), nil
}

// InsertData вставляет данные о заказе в базу данных.
func (s *Storage) InsertData(conn *pgx.Conn, order model.OrderInfo) {
	err := s.InsertPaymentData(conn, order)
	if err != nil {
		fmt.Println(err)
	}
	uid, err := s.InsertOrderData(conn, order)
	if err != nil {
		fmt.Println(err)
	}
	id, err := s.InsertDeliveryData(conn, order)
	if err != nil {
		fmt.Println(err)
	}
	err = s.InsertOrderDelivery(conn, uid, id)
	if err != nil {
		fmt.Println(err)
	}
	err = s.InsertItemsData(conn, order)
	if err != nil {
		fmt.Println(err)
	}
}

// InsertOrderData вставляет информацию о заказе в базу данных.
func (s *Storage) InsertOrderData(conn *pgx.Conn, order model.OrderInfo) (order_uid string, err error) {
	query := `
		insert into order_info
			(order_uid, track_number, entry, locale, internal_signature,customer_id, delivery_service, shardkey, sm_id, date_created,oof_shard) 
		values 
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		returning order_uid
		`
	if err = conn.QueryRow(context.Background(), query,
		order.OrderUid,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerId,
		order.DeliveryService,
		order.Shardkey,
		order.SmId,
		order.DateCreated,
		order.OofShard,
	).Scan(&order_uid); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			return "", pgErr
		}
	}
	return order_uid, nil
}

// InsertDeliveryData вставляет информацию о доставке в базу данных.
func (s *Storage) InsertDeliveryData(conn *pgx.Conn, order model.OrderInfo) (id int, err error) {
	query := `
		insert into deliveries
			(name, phone, zip, city, address, region, email) 
		values 
			($1,$2,$3,$4,$5,$6,$7)
		returning id
		`
	if err = conn.QueryRow(context.Background(), query,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	).Scan(&id); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return 0, pgErr
		}
	}
	return id, nil
}

// InsertPaymentData вставляет информацию о платеже в базу данных.
func (s *Storage) InsertPaymentData(conn *pgx.Conn, order model.OrderInfo) (err error) {
	query := `
		insert into payments
			(transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total,
                      custom_fee) 
		values 
			($1,$2,$3,$4,$5,$6,$7, $8, $9, $10)
		`
	if err = conn.QueryRow(context.Background(), query,
		order.Payment.Transaction,
		order.Payment.RequestId,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
		// Передача параметров для запроса.

	).Scan(); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return pgErr
		}
	}
	return nil
}

// InsertOrderDelivery вставляет информацию о связи заказа и доставки в базу данных.
func (s *Storage) InsertOrderDelivery(conn *pgx.Conn, order_uid string, id int) (err error) {
	query := `
		insert into order_delivery
			(order_uid,delivery_id)
		values 
			($1,$2)
		`
	if err = conn.QueryRow(context.Background(), query, order_uid, id).Scan(); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return pgErr
		}
	}
	return nil
}

// InsertItemsData вставляет информацию о товарах в базу данных.
func (s *Storage) InsertItemsData(conn *pgx.Conn, order model.OrderInfo) (err error) {
	for i := 0; i < len(order.Items); i++ {
		query := `
			insert into items
				(chrt_id,track_number,price,rid,name,sale,size,total_price,nm_id,brand,status)
			values
				($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			`
		if err = conn.QueryRow(context.Background(), query,
			order.Items[i].ChrtId,
			order.Items[i].TrackNumber,
			order.Items[i].Price,
			order.Items[i].Rid,
			order.Items[i].Name,
			order.Items[i].Sale,
			order.Items[i].Size,
			order.Items[i].TotalPrice,
			order.Items[i].NmId,
			order.Items[i].Brand,
			order.Items[i].Status,
		).Scan(); err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok {
				fmt.Println(pgErr)
				return pgErr
			}
		}
	}

	return nil
}
