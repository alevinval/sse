.PHONY: build
build:
	go build ./cmd/sse-consume

.PHONY: clean
clean:
	go clean ./...
	rm coverage.out
	rm sse-consume

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: cover
cover:
	go test -count=1 -cover -coverprofile coverage.out ./...
	go tool cover -html coverage.out

.PHONY: bench-old
bench-old:
	go test -cpu=8 -benchtime=1s -bench=. -benchmem -count=5 ./... > old.txt

.PHONY: bench-new
bench-new:
	go test -cpu=8 -benchtime=1s -bench=. -benchmem -count=5 ./... > new.txt

.PHONY: bench-compare
bench-compare:
	benchstat old.txt new.txt

.PHONY: mod-update
mod-update:
	go get -u ./...
	go mod tidy

git_dirty := $(shell git status -s)

.PHONY: git-clean-check
git-clean-check:
ifneq ($(git_dirty),)
	@echo "Git repository is dirty!"
	@false
else
	@echo "Git repository is clean."
endif

.PHONY: format
format:
	yq -i .github/workflows/ci.yml
	for _file in $$(gofmt -s -l . | grep -vE '^vendor/'); do \
		gofmt -s -w $$_file ; \
	done

.PHONY: format-check
format-check:
ifneq ($(git_dirty),)
	$(error format-check must be invoked on a clean repository)
endif
	$(MAKE) format
	$(MAKE) git-clean-check

.PHONY: authors
authors:
	scripts/generate-authors.sh
