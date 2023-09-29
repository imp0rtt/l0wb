# l0wb


Этот проект представляет собой сервис, разработанный на языке программирования Go. Проект включает в себя HTTP-сервер, взаимодействие с базой данных PostgreSQL, кэширование данных и обмен сообщениями через NATS Streaming.

## Описание

### Компоненты проекта

1. **HTTP-сервер:** Проект запускает HTTP-сервер, который обслуживает HTTP-запросы и предоставляет доступ к различным функциональным возможностям магазина. Маршруты HTTP обрабатываются с использованием библиотеки `github.com/julienschmidt/httprouter`.

2. **Хранилище данных (Storage):** Для хранения данных о заказах и других сущностях проект использует базу данных PostgreSQL. Взаимодействие с базой данных реализовано с помощью библиотеки `github.com/jackc/pgx/v4`.

3. **Кэш (Cache):** Для оптимизации доступа к данным используется кэширование. Данные о заказах извлекаются из базы данных и кэшируются, что позволяет уменьшить нагрузку на базу данных и ускорить доступ к данным. Кэш создается с использованием библиотеки `github.com/patrickmn/go-cache`.

4. **NATS Streaming:** Проект взаимодействует с NATS Streaming, службой для обмена сообщениями. Он подписывается на канал NATS Streaming и обрабатывает сообщения, которые поступают в этот канал.

5. **Docker-контейнеры:** Для удобства развертывания окружения для разработки и тестирования используются Docker-контейнеры. В проекте используются контейнеры Adminer (веб-интерфейс для управления базами данных) и PostgreSQL (база данных).


