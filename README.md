# Веб-планировщик задач

Дипломный проект: планировщик задач на Go с поддержкой повторяющихся событий,
упакованный в Docker.

## Возможности

- Создание, редактирование и удаление задач.
- Отметка задач выполненными (с автопереносом на следующую дату при повторе)
- Правила повторения: `d N` (каждые N дней), `w N,M` (дни недели),
  `m N,M` (дни месяца), `y` (ежегодно)
- Поиск по тексту и по дате.
- Опциональная аутентификация по паролю (JWT, HS256)

## Стек

- **Go** 1.26.1
- **SQLite** (драйвер `modernc.org/sqlite`, без CGO)
- **JWT** (`github.com/golang-jwt/jwt/v5`)
- **Docker** + Docker Compose

## Запуск через Docker

### Вариант 1 — `docker run`
```bash
docker run -p 7540:7540 -v ${PWD}/data:/app/data -e SALT_JWT=<твой-секрет> pyromeow/task-scheduler:latest
```
Для cmd используй `%cd%` вместо `${PWD}`, для bash — `$(pwd)`.
### Вариант 2 — `docker compose`
#### 1. Скачать образ (Образ отправлен в Docker HUB)
Образ опубликован https://hub.docker.com/repository/docker/pyromeow/task-scheduler/general
```bash
docker pull pyromeow/task-scheduler:latest
``` 
#### 2. Создать файл `.env` рядом с проектом
```
SALT_JWT=любая-случайная-строка-минимум-32-символа
```
#### 3. Создать `docker-compose.yml` (Если его нет)
```yaml
services:
  scheduler:
    image: pyromeow/task-scheduler:latest
    ports:
      - "7540:7540"
    volumes:
      - ./data:/app/data
    environment:
      - SALT_JWT=${SALT_JWT}
```
#### 4. Запустить
```bash
docker compose up
```
#### 5. Открыть в браузере
http://localhost:7540/

База данных `scheduler.db` появится в папке `./data` на хосте.

## Запуск из исходников

### Требования

- Go 1.26+
- Docker для запуска контейнера (необязательно)

### Шаги

```bash
git clone https://github.com/pyromeowww/task-scheduler.git
cd task-scheduler
go mod download

# создай .env с секретом
echo "SALT_JWT=$(openssl rand -hex 32)" > .env

go run .
```

Открой http://localhost:7540/

## Переменные окружения

| Переменная      | Обязательна | По умолчанию   | Описание                                       |
|-----------------|-------------|----------------|------------------------------------------------|
| `SALT_JWT`      | **да**      | —              | Секрет для подписи JWT-токенов                 |
| `TODO_PORT`     | нет         | `7540`         | Порт веб-сервера                               |
| `TODO_DBFILE`   | нет         | `scheduler.db` | Путь к файлу SQLite                            |
| `TODO_PASSWORD` | нет         | пусто          | Пароль для входа. Пусто → аутентификация off   |
`TODO_PASSWORD` по умолчанию не указано из-за тестов. По желанию можно указать переменную и проверить работу http://localhost:7540/login.html

## Аутентификация

Аутентификация включается автоматически, если задана переменная `TODO_PASSWORD`.

- **`TODO_PASSWORD` пуст** — все API открыты, страница `/login.html` вернёт
  «Аутентификация не настроена».
- **`TODO_PASSWORD` задан** — все API (кроме `/api/signin`) требуют
  валидный JWT-токен в cookie `token`.

Токен подписывается секретом `SALT_JWT` методом HS256. Время жизни — 8 часов.

## API

| Метод  | Endpoint              | Описание                                      |
|--------|-----------------------|-----------------------------------------------|
| POST   | `/api/signin`         | Аутентификация, возвращает `{"token": "..."}` |
| GET    | `/api/nextdate`       | Ближайшая дата по правилу повторения          |
| GET    | `/api/tasks`          | Список задач                                  |
| POST   | `/api/task`           | Создать задачу                                |
| GET    | `/api/task?id=?`      | Получить задачу по ID                         |
| PUT    | `/api/task`           | Обновить задачу                               |
| DELETE | `/api/task?id=?`      | Удалить задачу по ID                          |
| POST   | `/api/task/done?id=?` | Отметить задачу выполненной                   |
Все эндпоинты, кроме `/api/signin` и `/api/nextdate`, защищены middleware,
если задан `TODO_PASSWORD`.

## Тесты
```bash
go test ./tests
```

<details>
<summary>Список отдельных тестов (для отладки)</summary>

- `go test -run ^TestDB$ ./tests` — проверка работы с БД.
- `go test -run ^TestNextDate$ ./tests` — функция NextDate и `/api/nextdate`,для проверки запустите сервер и затем команду.
- `go test -run ^TestAddTask$ ./tests` — добавление задачи.
- `go test -run ^TestTasks$ ./tests` — поиск.
- `go test -run ^TestTask$ ./tests` — редактирование.
- `go test -run ^TestDone$ ./tests` — выполнение задачи.
- `go test -run ^TestDelTask$ ./tests` — удаление задачи.

По умолчанию тесты используют `../scheduler.db`. И считает, что файл с БД находится в родительской директории. Вы можете указать другой путь в значении переменной `DBFile` в файле `./tests/settings.go.`
Путь можно переопределить через `TODO_DBFILE` или в `tests/settings.go`.
Перед запуском убедись, что `TODO_PASSWORD` **не задан** — тесты ожидают
открытые эндпоинты.

</details>


## Задачи со звёздочкой
Выполнены все задания со звёздочкой

## Директории

- В директории `tests` — находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.
- Директория `web` — содержит файлы фронтенда.
- Директория `pkg` — содержит все нужные директории для работы сервера `API`, `db`, `server`.
- Директория `api` — содержит файлы с обработчиками API запросов.
- Директория `db` — содержит файлы с кодом, отвечающие за работу БД.
- Директория `server` — содержит файл для запуска сервера.
- Директория `data` — содержит БД.
