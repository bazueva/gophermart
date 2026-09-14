include .env

lint:
	golangci-lint run

test:
	go test ./...

test-cover-html:
	go test -coverprofile=coverage.out ./...
	@cat coverage.out | grep -v -f .coverignore > coverage_filtered.out
	@echo "📊 Покрытие по функциям:"
	@go tool cover -func=coverage_filtered.out
	@echo ""
	@echo "📊 Общее покрытие:"
	@go tool cover -func=coverage_filtered.out | grep total
	go tool cover -html=coverage_filtered.out -o coverage.html
	@echo "📊 HTML отчет сохранен в coverage.html"
	@open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || echo "Откройте файл в браузере: coverage.html"
	@rm -f coverage.out coverage_filtered.out

.PHONY: jet-generate
jet-generate:
	jet -dsn="$(DATABASE_DSN)" \
		-schema=public \
		-path=./schema.gen \
		-ignore-tables="schema_migrations"