build:
	go build -o ./bin/main ./src/main.go ./src/safeQueue.go ./src/types.go ./src/chatNode.go ./src/rateLimiter.go

run:
	go run ./src/main.go ./src/safeQueue.go ./src/types.go ./src/chatNode.go ./src/rateLimiter.go

clean:
	rm -rf bin/*

run_build:
	./bin/main

install:
	go install ./src/main.go ./src/safeQueue.go ./src/types.go ./src/chatNode.go ./src/rateLimiter.go