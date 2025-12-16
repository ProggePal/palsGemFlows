build:
	go build -o pals-gemflows ./cmd/pals-gemflows

test:
	go test ./...

release:
	sh ./scripts/package.sh
