# ============================================================
# Accountant
# ============================================================

APP_NAME        := accountant
CMD_DIR         := ./cmd/accountant
BIN_DIR         := ./bin

# gogpu/ui is maintained as a Git submodule.
UI_DIR          := ./pkg/gogpu-ui
UI_REPO         := https://github.com/gogpu/ui.git

# Git
BRANCH          := $(shell git branch --show-current)
COMMIT          := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION         := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go
GO              := go
GOFMT           := gofmt

# Build metadata
BUILD_TIME      := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "\
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildTime=$(BUILD_TIME)"

# ============================================================
# Help
# ============================================================

.PHONY: help
help:
	@echo ""
	@echo "Accountant Development Commands"
	@echo "================================"
	@echo ""
	@echo "Application"
	@echo "  make run             Run application"
	@echo "  make build           Build application"
	@echo "  make install         Install application"
	@echo ""
	@echo "Code"
	@echo "  make fmt             Format Go source"
	@echo "  make vet             Run go vet"
	@echo "  make test            Run tests"
	@echo "  make test-race       Run tests with race detector"
	@echo "  make coverage        Generate test coverage"
	@echo "  make tidy            Run go mod tidy"
	@echo "  make check           Format + vet + test + build"
	@echo ""
	@echo "gogpu/ui"
	@echo "  make ui-status       Show submodule status"
	@echo "  make ui-update       Update gogpu/ui"
	@echo "  make ui-init         Initialize submodule"
	@echo "  make ui-log          Show gogpu/ui history"
	@echo "  make ui-branch       Show gogpu/ui branch"
	@echo ""
	@echo "Git"
	@echo "  make git-status      Show Git status"
	@echo "  make git-log         Show recent commits"
	@echo "  make git-diff        Show unstaged changes"
	@echo "  make git-diff-cached Show staged changes"
	@echo "  make git-branches    List branches"
	@echo "  make git-tags        List tags"
	@echo "  make git-pull        Pull latest changes"
	@echo "  make git-push        Push current branch"
	@echo "  make git-sync        Pull + update submodules"
	@echo ""
	@echo "Git workflow"
	@echo "  make commit MSG='message'"
	@echo "  make push"
	@echo "  make pull"
	@echo "  make branch NAME=feature/name"
	@echo "  make switch NAME=branch"
	@echo "  make merge NAME=branch"
	@echo "  make tag VERSION=v0.1.0"
	@echo "  make release VERSION=v0.1.0"
	@echo ""
	@echo "Cleanup"
	@echo "  make clean           Remove build artifacts"
	@echo ""
	@echo "Information"
	@echo "  make info            Show project information"
	@echo ""


# ============================================================
# Application
# GOGPU_GRAPHICS_API=software GOGPU_RENDER_MODE=cpu make run
# ============================================================

.PHONY: run
run:
	$(GO) run $(CMD_DIR)

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

.PHONY: install
install:
	$(GO) install $(CMD_DIR)


# ============================================================
# Code Quality
# ============================================================

.PHONY: fmt
fmt:
	@echo "Formatting Go source..."
	$(GOFMT) -w .

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test ./...

.PHONY: test-race
test-race:
	$(GO) test -race ./...

.PHONY: coverage
coverage:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html: coverage
	$(GO) tool cover -html=coverage.out

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: check
check: fmt tidy vet test build
	@echo ""
	@echo "✓ All checks passed"
	@echo ""


# ============================================================
# Dependencies
# ============================================================

.PHONY: deps
deps:
	$(GO) mod download

.PHONY: deps-update
deps-update:
	$(GO) get -u ./...
	$(GO) mod tidy


# ============================================================
# gogpu/ui Submodule
# ============================================================

.PHONY: ui-init
ui-init:
	@echo "Initializing gogpu/ui submodule..."
	git submodule update --init --recursive

.PHONY: ui-status
ui-status:
	@echo ""
	@echo "gogpu/ui"
	@echo "---------"
	@git -C $(UI_DIR) status
	@echo ""
	@echo "Commit:"
	@git -C $(UI_DIR) log -1 --oneline
	@echo ""
	@echo "Branch:"
	@git -C $(UI_DIR) branch --show-current
	@echo ""

.PHONY: ui-update
ui-update:
	@echo "Updating gogpu/ui..."
	git submodule update --remote $(UI_DIR)
	$(GO) mod tidy
	@echo ""
	@echo "gogpu/ui updated to:"
	@git -C $(UI_DIR) log -1 --oneline

.PHONY: ui-log
ui-log:
	git -C $(UI_DIR) log --oneline --decorate -20

.PHONY: ui-branch
ui-branch:
	git -C $(UI_DIR) branch -a

.PHONY: ui-fetch
ui-fetch:
	git -C $(UI_DIR) fetch --all --prune

.PHONY: ui-pull
ui-pull:
	git -C $(UI_DIR) pull --ff-only


# ============================================================
# Git - Information
# ============================================================

.PHONY: git-status
git-status:
	git status

.PHONY: git-log
git-log:
	git log --oneline --decorate --graph -20

.PHONY: git-diff
git-diff:
	git diff

.PHONY: git-diff-cached
git-diff-cached:
	git diff --cached

.PHONY: git-branches
git-branches:
	git branch -a

.PHONY: git-tags
git-tags:
	git tag --list --sort=-version:refname

.PHONY: git-remote
git-remote:
	git remote -v


# ============================================================
# Git - Remote Operations
# ============================================================

.PHONY: git-fetch
git-fetch:
	git fetch --all --prune

.PHONY: git-pull
git-pull:
	git pull --ff-only

.PHONY: git-push
git-push:
	git push

.PHONY: git-sync
git-sync:
	git pull --ff-only
	git submodule update --init --recursive


# Aliases
.PHONY: pull
pull: git-pull

.PHONY: push
push: git-push


# ============================================================
# Git - Working Tree
# ============================================================

.PHONY: add
add:
	git add .

.PHONY: unstage
unstage:
	git restore --staged .

.PHONY: discard
discard:
	@echo "WARNING: This will discard ALL unstaged changes."
	@read -p "Continue? [y/N] " answer; \
	if [ "$$answer" = "y" ]; then \
		git restore .; \
	else \
		echo "Cancelled."; \
	fi


# ============================================================
# Git - Commit
# ============================================================

.PHONY: commit
commit:
ifndef MSG
	$(error MSG is required. Example: make commit MSG="Add account model")
endif
	git add .
	git commit -m "$(MSG)"


# ============================================================
# Git - Branches
# ============================================================

.PHONY: branch
branch:
ifndef NAME
	$(error NAME is required. Example: make branch NAME=feature/accounts)
endif
	git switch -c $(NAME)

.PHONY: switch
switch:
ifndef NAME
	$(error NAME is required. Example: make switch NAME=main)
endif
	git switch $(NAME)

.PHONY: merge
merge:
ifndef NAME
	$(error NAME is required. Example: make merge NAME=feature/accounts)
endif
	git merge $(NAME)


# ============================================================
# Git - Tags / Releases
# ============================================================

.PHONY: tag
tag:
ifndef VERSION
	$(error VERSION is required. Example: make tag VERSION=v0.1.0)
endif
	git tag -a $(VERSION) -m "Release $(VERSION)"

.PHONY: tag-push
tag-push:
ifndef VERSION
	$(error VERSION is required. Example: make tag-push VERSION=v0.1.0)
endif
	git push origin $(VERSION)

.PHONY: release
release:
ifndef VERSION
	$(error VERSION is required. Example: make release VERSION=v0.1.0)
endif
	@echo "Preparing release $(VERSION)..."
	$(MAKE) check
	git diff --exit-code
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(BRANCH)
	git push origin $(VERSION)


# ============================================================
# Project Information
# ============================================================

.PHONY: info
info:
	@echo ""
	@echo "Project       : $(APP_NAME)"
	@echo "Branch        : $(BRANCH)"
	@echo "Commit        : $(COMMIT)"
	@echo "Version       : $(VERSION)"
	@echo "Go            : $$(go version)"
	@echo ""
	@echo "gogpu/ui:"
	@git -C $(UI_DIR) log -1 --oneline
	@echo ""


# ============================================================
# Cleanup
# ============================================================

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -f coverage.out