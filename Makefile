all: run

run:
	go run ./cmd/... -config=dev.yml

build:
	go build -o atomic-pilot ./cmd/...

test:
	go test -v ./...

clean:
	rm -r dist/ atomic-pilot || true

update:
	go get -u ./...
	go mod tidy
