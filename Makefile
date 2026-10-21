COMPOSE_FILE := deploy/compose/compose.yml

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	mkdir -p bin
	go build -o bin/iam-server ./cmd/iam-server
	go build -o bin/iam-migrate ./cmd/iam-migrate
	go build -o bin/iam-keygen ./cmd/iam-keygen

.PHONY: keygen
keygen:
	go run ./cmd/iam-keygen

.PHONY: compose-up
compose-up:
	docker compose --env-file .env -f $(COMPOSE_FILE) up --build -d

.PHONY: compose-down
compose-down:
	docker compose --env-file .env -f $(COMPOSE_FILE) down

.PHONY: compose-logs
compose-logs:
	docker compose --env-file .env -f $(COMPOSE_FILE) logs -f

.PHONY: compose-ps
compose-ps:
	docker compose --env-file .env -f $(COMPOSE_FILE) ps

.PHONY: clean-data
clean-data:
	@echo "This deletes the local PostgreSQL volume."
	@echo "Run the following manually if that is really intended:"
	@echo "docker compose --env-file .env -f $(COMPOSE_FILE) down --volumes"