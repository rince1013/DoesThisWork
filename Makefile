APP      := doesthiswork
BUILD    := ./$(APP)
LINUX    := GOOS=linux GOARCH=arm64

DEPLOY_HOST ?= your-vps-ip
DEPLOY_USER ?= ubuntu
REMOTE      := $(DEPLOY_USER)@$(DEPLOY_HOST)
REMOTE_DIR  := /opt/doesthiswork

.PHONY: run build build-linux deploy logs clean

run:
	go run . serve --http=localhost:8090

build:
	go build -o $(BUILD) .

build-linux:
	$(LINUX) go build -ldflags="-s -w" -o $(BUILD) .

deploy: build-linux
	ssh $(REMOTE) "mkdir -p $(REMOTE_DIR)/pb_data $(REMOTE_DIR)/static"
	scp $(BUILD) $(REMOTE):$(REMOTE_DIR)/doesthiswork
	scp -r static/ $(REMOTE):$(REMOTE_DIR)/static/
	scp -r migrations/ $(REMOTE):$(REMOTE_DIR)/migrations/ 2>/dev/null || true
	ssh $(REMOTE) "sudo systemctl restart doesthiswork"
	@echo "Deployed to $(DEPLOY_HOST)"

logs:
	ssh $(REMOTE) "sudo journalctl -u doesthiswork -f --no-pager"

clean:
	rm -f $(BUILD)
