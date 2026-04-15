APP   := doesthiswork
BUILD := ./$(APP)
GO    := mise exec -- go

.PHONY: run build deploy clean

run:
	$(GO) run . serve --http=localhost:8090

build:
	$(GO) build -o $(BUILD) .

deploy:
	fly deploy

clean:
	rm -f $(BUILD)
