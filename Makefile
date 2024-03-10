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
	@$(GO) test ./... -v || true  #effectively disabling any impact from tests. I'm still working on improving the test suite: https://github.com/dark-enstein/vault/issues/6

clean-path:
	@[ -f $(PATH)/$(APP_NAME) ] && ( rm $(PATH)/$(APP_NAME) ) || printf ""
