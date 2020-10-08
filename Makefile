fmt:
	goimports -l -w .
	go mod tidy
	terraform fmt --recursive
