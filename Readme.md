В `.env` задаются переменные:

* `DATABASE_URL`
* `SERVER_PORT`
* `AGIFY_URL`
* `GENDERIZE_URL`
* `NATIONALIZE_URL`
* `ZAP_LEVEL`

Миграции применяются с помощью golang-migrate.

Установка golang-migrate:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Применение миграций:

```bash
migrate -path migrations -database "postgres://effectiveMobile:effectiveMobile@localhost:5432/effectiveMobile?sslmode=disable" up
```

Запуск приложения:

```bash
go run cmd/apiserver/main.go
```
