TARGETS := $(shell ls scripts)

$(TARGETS):
	@./scripts/$@

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS)
