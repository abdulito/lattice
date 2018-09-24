# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

# build
.PHONY: all
all: gazelle \
     build

.PHONY: build
build: PLATFORM ?=
build: TARGET ?= //cmd/... //pkg/...
build:
	@bazel build \
		$(PLATFORM) \
		$(TARGET)

.PHONY: build.platform.all
build.platform.all: build.platform.darwin \
                    build.platform.linux

.PHONY: build.platform.darwin
build.platform.darwin: gazelle
	@$(MAKE) build.no-gazelle PLATFORM=--platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64

.PHONY: build.platform.linux
build.platform.linux: gazelle
	@$(MAKE) build.no-gazelle PLATFORM=--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

.PHONY: clean
clean:
	@bazel clean


# testing
.PHONY: test
test: TARGET ?= //pkg/...
test: OUTPUT ?= errors
test: ARGS ?=
test: gazelle
	@bazel test \
		$(ARGS) \
		--test_output=$(OUTPUT) \
		$(TARGET)

.PHONY: test.no-cache
test.no-cache:
	@$(MAKE) test ARGS=--cache_test_results=no

.PHONY: test.verbose
test.verbose:
	@$(MAKE) test OUTPUT=all ARGS="--test_env -v"


# formatting/linting
.PHONY: check
check: gazelle \
       format  \
       vet     \
       lint.no-export-comments

.PHONY: format
format:
	@gofmt -w .
	@terraform fmt .

.PHONY: lint
lint: install.golint
	@golint ./... | grep -v "customresource/generated" | grep -v "zz_generated."

.PHONY: lint.no-export-comments
lint.no-export-comments: install.golint
	@$(MAKE) lint | grep -v " or be unexported" | grep -v "comment on exported "

.PHONY: vet
vet: install.govet
	@go tool vet .


# tool installation
.PHONY: install.golint
install.golint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then go get github.com/golang/lint/golint; fi

.PHONY: install.govet
install.govet:
	@go tool vet 2>/dev/null; if [ $$? -eq 3 ]; then go get golang.org/x/tools/cmd/vet; fi


# git
.PHONY: git.install-hooks
git.install-hooks:
	cp -f hack/git/pre-commit.sh $(DIR)/.git/hooks/pre-commit
	cp -f hack/git/pre-push.sh $(DIR)/.git/hooks/pre-push


# docgen
.PHONY: docgen.latticectl
docgen.latticectl:
	@bazel run //cmd/latticectl:docgen -- --plugin cmd/latticectl/plugin.so


# local
.PHONY: local.up
local.up: CHANNEL=gcr.io/lattice-dev
local.up:
	@$(DIR)/hack/local/up.sh --set containerChannel=$(CHANNEL)

.PHONY: local.down
local.down:
	@$(DIR)/hack/local/down.sh


# docker
.PHONY: docker.build
docker.build:
	@bazel build //docker/...

.PHONY: docker.all.push
docker.all.push: docker.kubernetes.all.push \
                 docker.mock.all.push

KUBERNETES_DOCKER_IMAGES := api-server                  \
                            container-builder           \
                            envoy-prepare               \
                            envoy-xds-api               \
                            installer-helm              \
                            controller-manager          \
                            local-dns-controller

KUBERNETES_IMAGE_TARGETS := $(addprefix docker.kubernetes.,$(KUBERNETES_DOCKER_IMAGES))

.PHONY: docker.kubernetes.push.all
docker.kubernetes.all.push:
	@for image in $(KUBERNETES_DOCKER_IMAGES); do \
		$(MAKE) docker.kubernetes.$$image.push || exit 1; \
	done

KUBERNETES_IMAGE_PUSHES := $(addsuffix .push,$(KUBERNETES_IMAGE_TARGETS))
.PHONY: $(KUBERNETES_IMAGE_PUSHES)
$(KUBERNETES_IMAGE_PUSHES):
	@$(MAKE) docker.push \
    		TARGET_DIR=/kubernetes \
    		TARGET=$(patsubst docker.kubernetes.%.push,%,$@)

KUBERNETES_STRIPPED_IMAGE_PUSHES := $(addsuffix .push.stripped,$(KUBERNETES_IMAGE_TARGETS))
.PHONY: $(KUBERNETES_STRIPPED_IMAGE_PUSHES))
$(KUBERNETES_STRIPPED_IMAGE_PUSHES):
	@$(MAKE) docker.push.stripped \
    		TARGET_DIR=/kubernetes \
    		TARGET=$(patsubst docker.kubernetes.%.push.stripped,%,$@)

KUBERNETES_IMAGE_RUNS := $(addsuffix .run,$(KUBERNETES_IMAGE_TARGETS))
.PHONY: $(KUBERNETES_IMAGE_RUNS)
$(KUBERNETES_IMAGE_RUNS):
	@$(MAKE) docker.run \
    		TARGET_DIR=/kubernetes \
    		TARGET=$(patsubst docker.kubernetes.%.run,%,$@)

KUBERNETES_IMAGE_SAVES := $(addsuffix .save,$(KUBERNETES_IMAGE_TARGETS))
.PHONY: $(KUBERNETES_IMAGE_SAVES)
$(KUBERNETES_IMAGE_SAVES):
	@$(MAKE) docker.save \
		TARGET_DIR=/kubernetes \
		TARGET=$(patsubst docker.kubernetes.%.save,%,$@)

KUBERNETES_IMAGE_SHS := $(addsuffix .sh,$(KUBERNETES_IMAGE_TARGETS))
.PHONY: $(KUBERNETES_IMAGE_SHS)
$(KUBERNETES_IMAGE_SHS):
	@$(MAKE) docker.sh \
    		TARGET_DIR=/kubernetes \
    		TARGET=$(patsubst docker.kubernetes.%.sh,%,$@)

MOCK_DOCKER_IMAGES := api-server

MOCK_IMAGE_TARGETS := $(addprefix docker.mock.,$(MOCK_DOCKER_IMAGES))

.PHONY: docker.mock.push.all
docker.mock.all.push:
	@for image in $(MOCK_DOCKER_IMAGES); do \
		$(MAKE) docker.mock.$$image.push || exit 1; \
	done

MOCK_IMAGE_PUSHES := $(addsuffix .push,$(MOCK_IMAGE_TARGETS))
.PHONY: $(MOCK_IMAGE_PUSHES)
$(MOCK_IMAGE_PUSHES):
	@$(MAKE) docker.push \
    		TARGET_DIR=/mock \
    		TARGET=$(patsubst docker.mock.%.push,%,$@)

MOCK_STRIPPED_IMAGE_PUSHES := $(addsuffix .push.stripped,$(MOCK_IMAGE_TARGETS))
.PHONY: $(MOCK_STRIPPED_IMAGE_PUSHES))
$(MOCK_STRIPPED_IMAGE_PUSHES):
	@$(MAKE) docker.push.stripped \
    		TARGET_DIR=/mock \
    		TARGET=$(patsubst docker.mock.%.push.stripped,%,$@)

MOCK_IMAGE_RUNS := $(addsuffix .run,$(MOCK_IMAGE_TARGETS))
.PHONY: $(MOCK_IMAGE_RUNS)
$(MOCK_IMAGE_RUNS):
	@$(MAKE) docker.run \
    		TARGET_DIR=/mock \
    		TARGET=$(patsubst docker.mock.%.run,%,$@)

MOCK_IMAGE_SAVES := $(addsuffix .save,$(MOCK_IMAGE_TARGETS))
.PHONY: $(MOCK_IMAGE_SAVES)
$(MOCK_IMAGE_SAVES):
	@$(MAKE) docker.save \
		TARGET_DIR=/mock \
		TARGET=$(patsubst docker.mock.%.save,%,$@)

MOCK_IMAGE_SHS := $(addsuffix .sh,$(MOCK_IMAGE_TARGETS))
.PHONY: $(MOCK_IMAGE_SHS)
$(MOCK_IMAGE_SHS):
	@$(MAKE) docker.sh \
    		TARGET_DIR=/mock \
    		TARGET=$(patsubst docker.mock.%.sh,%,$@)


.PHONY: docker.push
docker.push:
	@bazel run \
		--workspace_status_command "REGISTRY=$(REGISTRY) CHANNEL=$(CHANNEL) $(DIR)/hack/bazel/docker-workspace-status.sh" \
		//docker$(TARGET_DIR):push-$(TARGET)

.PHONY: docker.push.stripped
docker.push.stripped:
	@bazel run \
		--workspace_status_command "REGISTRY=$(REGISTRY) CHANNEL=$(CHANNEL) $(DIR)/hack/bazel/docker-workspace-status.sh" \
		//docker$(TARGET_DIR):push-$(TARGET)-stripped

.PHONY: docker.run
docker.run:
	@bazel run //docker$(TARGET_DIR):$(TARGET)

.PHONY: docker.save
docker.save:
	@bazel run //docker$(TARGET_DIR):$(TARGET) -- --norun

.PHONY: docker.sh
docker.sh: docker.save
	docker run -it --entrypoint sh bazel/docker$(TARGET_DIR):$(TARGET)

# kubernetes
.PHONY: kubernetes.update-dependencies
kubernetes.update-dependencies:
	LATTICE_ROOT=$(DIR) KUBERNETES_VERSION=$(VERSION) $(DIR)/hack/kubernetes/dependencies/update-kubernetes-version.sh
	$(MAKE) kubernetes.regenerate-custom-resource-clients VERSION=$(VERSION)

.PHONY: kubernetes.regenerate-custom-resource-clients
kubernetes.regenerate-custom-resource-clients:
	KUBERNETES_VERSION=$(VERSION) $(DIR)/hack/kubernetes/codegen/regenerate.sh
	@$(MAKE) gazelle
