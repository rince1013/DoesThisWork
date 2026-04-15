APP   := doesthiswork
BUILD := ./$(APP)
GO    := mise exec -- go

.PHONY: run build deploy clean

run:
	$(GO) run . serve --http=localhost:8090

build:
	$(GO) build -o $(BUILD) .

deploy:
	git push origin main
	@echo "Render will build and deploy automatically."

clean:
	rm -f $(BUILD)
