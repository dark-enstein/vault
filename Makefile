all: build

.PHONY: build sudo clean-path test
CMD_DIR="vaught"
PATH="/usr/local/bin"
APP_NAME="vault"
GO=$(shell which go)

build:
	@$(GO) build -o $(APP_NAME) $(CMD_DIR)/main.go

sudo:
	@$(GO) build -o $(APP_NAME) $(CMD_DIR)/main.go
	@sudo mv vault $(PATH)

test:
	@$(GO) test ./... -v

clean-path:
	@[ -f $(PATH)/$(APP_NAME) ] && ( rm $(PATH)/$(APP_NAME) ) || printf ""
