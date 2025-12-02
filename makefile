export GO111MODULE=on

help: ## This help dialog.
help h:
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##/:/'`); \
	printf "%-20s %s\n" "target" "help" ; \
	printf "%-20s %s\n" "------" "----" ; \
	for help_line in $${help_lines[@]}; do \
		IFS=$$':' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf '\033[36m'; \
		printf "%-20s %s" $$help_command ; \
		printf '\033[0m'; \
		printf "%s\n" $$help_info; \
	done

tests:
	@echo "=================="
	@echo "Running unit tests"
	@echo "=================="
	go test -tags unit -shuffle=on -coverprofile coverage.out ./...


format:
	@echo "=========================================="
	@echo "Formatting your code"
	@echo "=========================================="
	gci write -s standard -s default . --skip-generated --skip-vendor  && gofumpt -l -w .



lint: ## Run all enabled linters
	@echo "=========================================="
	@echo "Running static analysis"
	@echo "Use 'fix=true' to fix issues automatically"
	@echo "Use 'use_github_pat=true' to use github personal access token to download packages"
	@echo "=========================================="
	LINTER_FLAGS="-v"; \
    	golangci-lint run ${LINTER_FLAGS}; \
    	docker run -t --rm -v ${PWD}:/app -v $$(go env GOMODCACHE):/go/pkg/mod  -w /app golangci/golangci-lint:v2.6.2 sh -c "$${CMD}"olangci/golangci-lint:v1.61.0 sh -c "$${CMD}"