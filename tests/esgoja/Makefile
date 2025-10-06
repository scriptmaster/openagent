.PHONY: test stop start tail vet

PORT := 8802
BIN := esgoja-test
LOG := /tmp/$(BIN).log

# Build + run server in background on 8802, after killing any binder on that port
# Conforms to user's note: use `make test &` to run in background

test: stop build run
	@echo "Server listening at http://localhost:$(PORT)"
	@echo "Logs: tail -f $(LOG)"

build:
	@echo "Tidying modules..."
	@cd $(CURDIR) && go mod tidy
	@echo "Building..."
	@cd $(CURDIR) && go build -o $(BIN) ./

run:
	@echo "Starting on :$(PORT) ..."
	@nohup ./$(BIN) > $(LOG) 2>&1 &

stop:
	@echo "Stopping anything on port $(PORT)..."
	@lsof -ti:$(PORT) | xargs kill -9 2>/dev/null || true

start: test

tail:
	@tail -f $(LOG)

vet:
	@echo "Tidying modules..."
	@cd $(CURDIR) && go mod tidy
	@echo "Running go vet..."
	@cd $(CURDIR) && go vet ./...
