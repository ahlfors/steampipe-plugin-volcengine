STEAMPIPE_INSTALL_DIR ?= ~/.steampipe
BUILD_TAGS = netgo

build:
	go build -tags "${BUILD_TAGS}" -o ./ ./...

install:
	go build -o $(STEAMPIPE_INSTALL_DIR)/plugins/hub.steampipe.io/plugins/turbot/volcengine@latest/steampipe-plugin-volcengine.plugin -tags "${BUILD_TAGS}" *.go
