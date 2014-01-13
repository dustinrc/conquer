DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps
	@echo "==> Building"
	@go install


clean:
	@echo "==> Cleaning"
	@rm -rf cookies.txt conquer $(GOBIN)/conquer

deps:
	@echo "==> Installing dependencies"
	@go get -d -v ./...
	@echo $(DEPS) | xargs -n1 go get -d

updatedeps:
	@echo "==> Updating all dependencies"
	@go get -d -v -u ./...
	@echo $(DEPS) | xargs -n1 go get -d -u

test: all
	@echo "==> Testing..."
	@conquer &
	@bash test.sh whiskey hotel http://127.0.0.1:8080
	@killall conquer

.PHONY: all clean deps test updatedeps
