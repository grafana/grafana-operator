# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

DOCS_URL ?= http://localhost:1313

.PHONY: detect-broken-links
detect-broken-links:
	go install github.com/raviqqe/muffet/v2@latest
	muffet --rate-limit=3 \
		--max-connections=2 \
		--buffer-size=8192 \
		--exclude=github.com \
		--exclude=grafana.127.0.0.1.nip.io \
		$(DOCS_URL)

