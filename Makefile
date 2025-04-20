.PHONY: run postgres migrations_down test_unit test_e2e wait_for_postgres gen

postgres:
	docker run -d -p 5436:5432 -e POSTGRES_PASSWORD='qwerty' --name='postgres-db' postgres:17-alpine3.21

migrations_down:
	goose -dir internal/repository/postgres/migrations postgres "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" down

wait_for_postgres:
	until pg_isready -h localhost -p 5436 -U postgres; do \
		sleep 1; \
	done

run: postgres wait_for_postgres
	go run cmd/app/main.go --config=./config/config.yml

gen:
	mockgen -source=internal/service/auth/auth.go -destination=internal/service/auth/mocks/mock.go

test_e2e:
	go test ./tests/e2e/... -v

test_unit:
	go test ./tests/unit/... -v -count=1