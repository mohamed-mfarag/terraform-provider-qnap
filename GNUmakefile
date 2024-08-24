default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_CLI_ARGS_apply="-parallelism=1" TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
