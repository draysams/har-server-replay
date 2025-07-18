# Go project automation
.PHONY: test coverage

test:
	go test ./... -cover

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out 