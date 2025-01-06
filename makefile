PROJECT_NAME := "cloud-torrent-dler"
PKG := "github.com/coma-toast/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
SERVER := $(shell cat config.yaml | grep -m1 TargetServer | awk '{ print $$2 }')
SERVERDIR := $(shell cat config.yaml | grep TargetServerDir | awk '{ print $$2 }')

dep: ## Get the dependencies
	@go get -v -d ./...

build: dep ## Build the binary file
	@go build -v $(PKG)

build-mac: dep ## Build the binary file
	@env GOOS=darwin go build -v $(PKG)

install: ## Install the pre-req files and services so deploy will work
	ssh $(SERVER) "test -e $(SERVERDIR); or git clone git@gitlab.jasondale.me:jdale/$(PROJECT_NAME).git $(SERVERDIR); and git -C $(SERVERDIR) pull"
	ssh $(SERVER) "sudo cp $(SERVERDIR)/$(PROJECT_NAME).service /etc/systemd/system/$(PROJECT_NAME).service; and sudo chmod +x $(SERVERDIR)/$(PROJECT_NAME).sh; and sudo systemctl start $(PROJECT_NAME); and sudo systemctl enable $(PROJECT_NAME); and sudo systemctl daemon-reload"

deploy: build
	ssh $(SERVER) "sudo service $(PROJECT_NAME) stop"
	rsync $(PROJECT_NAME) $(SERVER):$(SERVERDIR)/
	rsync $(PROJECT_NAME).sh $(SERVER):$(SERVERDIR)/
	ssh $(SERVER) "sudo service $(PROJECT_NAME) start"

