all_go_files = $(shell find "." -name "*.go" -print)

ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)
OTEL_GO_MOD_DIRS := $(filter-out $(TOOLS_MOD_DIR), $(ALL_GO_MOD_DIRS))

go.work:
	go work init

.PHONY: setup
setup:
	go env -w GO111MODULE=on
	go env -w GOPROXY="https://goproxy.cn,direct"
	go env -w GOSUMDB="sum.golang.google.cn"

.PHONY: addlicense
addlicense:
	@addlicense -c "Beijing Volcano Engine Technology Co., Ltd." $(all_go_files)

.PHONY: gofmt
gofmt: $(OTEL_GO_MOD_DIRS:%=gofmt/%)
gofmt/%: DIR=$*
gofmt/%:
	@echo 'gofmt -w . && go mod tidy $(DIR)' \
		&& gofmt -w . && go mod tidy

.PHONY: unit-test
unit-test: $(OTEL_GO_MOD_DIRS:%=unit-test/%)
unit-test/%: DIR=$*
unit-test/%: go.work
	@echo 'go test -race -timeout 30s -gcflags="all=-l -N"-p 1 -parallel 1 -coverprofile=cover.out $(if $(ARGS),$(ARGS) ,)$(DIR)' \
		&& go work use $(DIR) \
		&& cd $(DIR) \
		&& go test -race -timeout 30s -gcflags="all=-l -N" -p 1 -parallel 1 -coverprofile=cover.out ./... \
		&& go tool cover -html=cover.out -o cover.html

.PHONY: golangci-lint
golangci-lint: $(OTEL_GO_MOD_DIRS:%=golangci-lint/%)
golangci-lint/%: DIR=$*
golangci-lint/%: go.work
	@echo 'golangci-lint run --new-from-rev=origin/main -v $(if $(ARGS),$(ARGS) ,)$(DIR)' \
		&& go work use $(DIR) \
		&& cd $(DIR) \
		&& golangci-lint run --allow-serial-runners $(ARGS)
