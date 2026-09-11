.PHONY: test grade vet fmt

test: vet
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

# Grade the reference solution.
grade:
	go run ./cmd/grade ./solutions/ch01

# Grade an arbitrary submission: make grade-dir DIR=/path/to/submission
grade-dir:
	go run ./cmd/grade $(DIR)
