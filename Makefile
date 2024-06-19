default: install

generate:
	go generate ./...

install:
	go install .

hooks:
	git config core.hooksPath .githooks

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...
