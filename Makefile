# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: build
build: gazelle
	@bazel build //...:all

.PHONY: clean
clean:
	@bazel clean

.PHONY: test
test: gazelle
	@bazel test --test_output=errors //...

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

.PHONY: local-up
local-up: minikube-start local-bootstrap

.PHONY: local-down
local-down: minikube-stop

.PHONY: local-delete
local-delete: minikube-delete

.PHONY: local-bootstrap
local-bootstrap:
	@bazel run -- //cmd/bootstrap -kubeconfig ~/.kube/config -logtostderr -provider local
	$(DIR)/bin/seed-local-build-images.sh

.PHONY: minikube-start
minikube-start:
	@minikube start

.PHONY: minikube-stop
minikube-stop:
	@minikube stop

.PHONY: minikube-delete
minikube-delete:
	@minikube delete
