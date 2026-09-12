.PHONY: test grade grade2 vet fmt

test: vet
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

# Grade the reference solutions.
grade:
	go run ./cmd/grade -ch 1 ./solutions/ch01

grade2:
	go run ./cmd/grade -ch 2 ./solutions/ch02

# Grade an arbitrary submission:
#   make grade-dir DIR=/path/to/submission        (chapter 1)
#   make grade-dir CH=2 DIR=/path/to/submission   (chapter 2)
CH ?= 1
grade-dir:
	go run ./cmd/grade -ch $(CH) $(DIR)
