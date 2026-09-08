PYTHON ?= python3
RELEASE := ./release.py
GO_DIR := go
LANGS := go
BUMP_PARTS := patch minor major

.PHONY: help setup hooks lint lint-fix test security check-release status bump $(LANGS) $(BUMP_PARTS)

help:
	@echo "Targets ($(notdir $(CURDIR))):"
	@echo "  make setup          — enable .githooks (run once per clone)"
	@echo "  make hooks          — git config core.hooksPath .githooks"
	@echo "  make lint           — go vet + gofmt check (go/)"
	@echo "  make lint-fix       — gofmt -w (go/)"
	@echo "  make test           — go test ./... (go/)"
	@echo "  make security       — govulncheck + gosec (go/)"
	@echo "  make check-release  — per-language VERSION + CHANGELOG gate vs main"
	@echo "  make status         — per-language version and tag state"
	@echo "  make bump <lang> patch|minor|major — bump <lang>/VERSION"

setup: hooks
	@echo "ready — local hooks enabled (.githooks). Bypass: git commit --no-verify"

hooks:
	git config core.hooksPath .githooks
	@chmod +x .githooks/* 2>/dev/null || true

lint:
	$(MAKE) -C $(GO_DIR) vet fmt-check

lint-fix:
	$(MAKE) -C $(GO_DIR) fmt

test:
	$(MAKE) -C $(GO_DIR) test

security:
	$(MAKE) -C $(GO_DIR) security

check-release:
	@base="$${RELEASE_BASE:-origin/main}"; \
	head="$${RELEASE_HEAD:-HEAD}"; \
	$(PYTHON) ci/gates.py mr "$$base" "$$head"

$(LANGS) $(BUMP_PARTS):
	@:

status:
	$(PYTHON) $(RELEASE) status

bump:
	@lang="$(filter $(LANGS),$(MAKECMDGOALS))"; \
	part="$(filter $(BUMP_PARTS),$(MAKECMDGOALS))"; \
	if [ -z "$$lang" ] || [ -z "$$part" ]; then \
		echo "usage: make bump <lang> patch|minor|major" >&2; \
		echo "langs: $(LANGS)" >&2; \
		exit 1; \
	fi; \
	$(PYTHON) $(RELEASE) bump $$lang $$part
