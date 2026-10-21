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
build: web-build
	mkdir -p bin
	go build -o bin/iam-server ./cmd/iam-server
	go build -o bin/iam-migrate ./cmd/iam-migrate
	go build -o bin/iam-keygen ./cmd/iam-keygen
	go build -o bin/iamctl ./cmd/iamctl
	go build -o bin/example-server ./examples/go-svelte/cmd/server

.PHONY: keygen
keygen:
	go run ./cmd/iam-keygen

.PHONY: compose-up
compose-up:
	docker-compose --env-file .env -f $(COMPOSE_FILE) up --build -d

.PHONY: compose-down
compose-down:
	docker-compose --env-file .env -f $(COMPOSE_FILE) down

.PHONY: compose-logs
compose-logs:
	docker-compose --env-file .env -f $(COMPOSE_FILE) logs -f

.PHONY: compose-ps
compose-ps:
	docker-compose --env-file .env -f $(COMPOSE_FILE) ps

.PHONY: admin-create
admin-create:
	@read -p "Administrator username: " username; \
	docker-compose --env-file .env -f $(COMPOSE_FILE) run \
		--rm \
		--entrypoint /app/iamctl \
		iam \
		admin create --username "$$username"

.PHONY: clean-data
clean-data:
	@echo "This deletes the local PostgreSQL volume."
	@echo "Run the following manually if that is really intended:"
	@echo "docker-compose --env-file .env -f $(COMPOSE_FILE) down --volumes"

.PHONY: web-install
web-install:
	cd web/admin && npm install

.PHONY: web-build
web-build:
	cd web/admin && npm run build

.PHONY: web-dev
web-dev:
	cd web/admin && npm run dev

.PHONY: example-build
example-build:
	go build ./examples/go-svelte/cmd/server

.PHONY: example-run
example-run:
	go run ./examples/go-svelte/cmd/server