build:
	go build -o cheflb cmd/cheflb/main.go

run: build
	./cheflb
