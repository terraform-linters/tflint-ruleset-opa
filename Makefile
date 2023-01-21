default: build

test:
	go test $$(go list ./... | grep -v integration)

e2e:
	go test ./integration

build:
	go build

install: build
	mkdir -p ~/.tflint.d/plugins
	mv ./tflint-ruleset-opa ~/.tflint.d/plugins
