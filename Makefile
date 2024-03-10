all: build

.PHONY: build sudo clean-path test
CMD_DIR="vaught"
PATH="/usr/local/bin"
APP_NAME="vault"

build:
	@go build -o $(APP_NAME) $(CMD_DIR)/main.go

sudo:
	@go build -o $(APP_NAME) $(CMD_DIR)/main.go
	@sudo mv vault $(PATH)

test:
	@go test ./...

clean-path:
	@[ -f $(PATH)/$(APP_NAME) ] && ( rm $(PATH)/$(APP_NAME) ) || printf ""
