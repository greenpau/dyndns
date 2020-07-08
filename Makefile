.PHONY: test ctest covdir coverage docs linter qtest clean dep release logo
APP="dyndns"
APP_VERSION:=$(shell cat VERSION | head -1)
GIT_COMMIT:=$(shell git describe --dirty --always)
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD -- | head -1)
LATEST_GIT_COMMIT:=$(shell git log --format="%H" -n 1 | head -1)
BUILD_USER:=$(shell whoami)
BUILD_DATE:=$(shell date +"%Y-%m-%d")
BUILD_DIR:=$(shell pwd)
VERBOSE:=-v
ifdef TEST
	TEST:="-run ${TEST}"
endif
BINARY=${APP}

all:
	@echo "Version: $(APP_VERSION), Branch: $(GIT_BRANCH), Revision: $(GIT_COMMIT)"
	@echo "Build on $(BUILD_DATE) by $(BUILD_USER)"
	@mkdir -p bin/
	@rm -rf ./bin/*
	@CGO_ENABLED=0 go build -o ./bin/$(BINARY) $(VERBOSE) \
		-ldflags="-w -s \
		-X main.appVersion=$(APP_VERSION) \
		-X main.gitBranch=$(GIT_BRANCH) \
		-X main.gitCommit=$(GIT_COMMIT) \
		-X main.buildUser=$(BUILD_USER) \
		-X main.buildDate=$(BUILD_DATE)" \
		-gcflags="all=-trimpath=$(GOPATH)/src" \
		-asmflags="all=-trimpath $(GOPATH)/src" \
		cmd/${APP}/*.go
	@./bin/$(BINARY) --validate --log-level debug
	@./bin/$(BINARY) --version

linter:
	@echo "Running lint checks"
	@golint -set_exit_status *.go
	@for f in `find ./ -type f -name '*.go'`; do echo $$f; go fmt $$f; golint -set_exit_status $$f; done
	@for f in `find ./assets -name *.y*ml`; do yamllint $$f; done
	@#cat assets/conf/config.yaml | yq . > assets/conf/config.json
	@echo "PASS: lint checks"

covdir:
	@echo "Creating .coverage/ directory"
	@mkdir -p .coverage

test: covdir linter
	@echo "Running go test"
	@go test $(VERBOSE) -coverprofile=.coverage/coverage.out ./*.go
	@echo "PASS: go test"

ctest: covdir linter
	@time richgo test $(VERBOSE) $(TEST) -coverprofile=.coverage/coverage.out ./*.go

coverage: covdir
	@echo "Running coverage"
	@go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html
	@go test -covermode=count -coverprofile=.coverage/coverage.out ./*.go
	@go tool cover -func=.coverage/coverage.out | grep -v "100.0"
	@echo "PASS: coverage"

docs:
	@mkdir -p .doc
	@go doc -all > .doc/index.txt

clean:
	@rm -rf .doc/ .coverage/ bin/ build/ pkg-build/
	@echo "OK: clean up completed"

qtest:
	@echo "Perform quick tests ..."
	@rm -rf .coverage/*
	@#go test $(VERBOSE) -coverprofile=.coverage/coverage.out -run TestServerConfig ./*.go
	@#./bin/$(BINARY) -config ./assets/conf/config.json -log-level debug

dep:
	@echo "Making dependencies check ..."
	@go get -u golang.org/x/lint/golint
	@go get -u golang.org/x/tools/cmd/godoc
	@go get -u github.com/greenpau/versioned/cmd/versioned
	@pip3 install yamllint --user
	@pip3 install yq --user

release:
	@echo "Making release"
	@if [ $(GIT_BRANCH) != "master" ]; then echo "cannot release to non-master branch $(GIT_BRANCH)" && false; fi
	@git diff-index --quiet HEAD -- || ( echo "git directory is dirty, commit changes first" && false )
	@versioned -patch
	@git add VERSION
	@git commit -m 'updated VERSION file'
	@versioned -sync cmd/${APP}/main.go
	@echo "Patched version"
	@git add cmd/${APP}/main.go
	@git commit -m "released v`cat VERSION | head -1`"
	@git tag -a v`cat VERSION | head -1` -m "v`cat VERSION | head -1`"
	@git push
	@git push --tags
	@@echo "If necessary, run the following commands:"
	@echo "  git push --delete origin v$(APP_VERSION)"
	@echo "  git tag --delete v$(APP_VERSION)"

logo:
	@mkdir -p assets/docs/images
	@convert -background black -fill white \
		-font DejaVu-Sans-Bold -size 640x320! \
		-gravity center -pointsize 96 label:'dyndns' \
		PNG32:assets/docs/images/logo.png
