build:
	@go build -o bin/beli-mang cmd/server/main.go

run: build
	@./bin/beli-mang