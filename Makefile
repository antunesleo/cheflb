build:
	go build -o cheflb cmd/server/main.go

run: build
	./cheflb
