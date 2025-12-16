build:
	go build -o my-tool ./cmd/my-tool

test:
	go test ./...

release:
	sh ./scripts/package.sh
