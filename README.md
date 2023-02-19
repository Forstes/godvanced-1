## godvanced-1

*Go + PostgreSQL JSON API*

### Download required packages before run:
`go get`

### Set .env variables before run:
1. Create .env (or .env.local) file in root directory: 
  `touch .env`
2. Set the following variables inside with your own values:
  DB_DSN=postgres://$USERNAME:$PASSWORD@localhost/$DB_NAME?sslmode=disable
  JWT_KEY=super_duper_key
  JWT_EXPIRY_HOURS=8

### How to run:
`go run ./cmd/api`

### Migrations:
1. Install [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) if you want to work with migrations:

2. Create your own migration files:
`migrate create -seq -ext=.sql -dir=./migrations $MIGRATION_NAME`

3. Apply your migration:
`migrate -path=./migrations -database="$DB_PATH" up`
