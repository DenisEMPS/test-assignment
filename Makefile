.PHONY: build run postgres migrations_up migrations_down test_unit test_e2e wait_for_postgres gen

build:
	go build -o app cmd/app/main.go

postgres:
	docker run -d -p 5436:5432 -e POSTGRES_PASSWORD='qwerty' --name='postgres' postgres:17-alpine3.21

migrations_up:
	goose -dir migrations postgres "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" up

migrations_down:
	goose -dir migrations postgres "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" down

wait_for_postgres:
	until pg_isready -h localhost -p 5436 -U postgres; do \
		sleep 1; \
	done

run: postgres wait_for_postgres migrations_up
	go run cmd/app/main.go --config=./config/config.yml

gen:
	mockgen -source=internal/service/auth/auth.go -destination=internal/service/auth/mocks/mock.go

test_e2e:
	go test ./tests/e2e/... -v

test_unit:
	go test ./tests/unit/... -v -count=1