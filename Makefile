# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CLOUD_IMAGE_DIR = $(DIR)/cloud-images
CLOUD_IMAGE_BUILD_DIR = $(CLOUD_IMAGE_DIR)/build
CLOUD_IMAGE_BUILD_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/build
CLOUD_IMAGE_AWS_SYSTEM_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/aws/$(LATTICE_SYSTEM_ID)

LOCAL_REGISTRY = lattice-local
DEV_REGISTRY = gcr.io/lattice-dev
DEV_TAG ?= latest

CONTAINER_NAME_BUILD = lattice-system-builder

BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH = debian-with-ssh
BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH_DEV = $(DEV_REGISTRY)/$(BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH):$(DEV_TAG)

BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES = debian-with-iptables
BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES_DEV = $(DEV_REGISTRY)/$(BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES):$(DEV_TAG)

BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS = ubuntu-with-aws
BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS_DEV = $(DEV_REGISTRY)/$(BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS):$(DEV_TAG)

# Basic build/clean/test
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

.PHONY: docker-build-base-images
docker-build-base-images:
	docker build $(DIR)/docker/component-build -f $(DIR)/docker/component-build/Dockerfile.aws -t $(BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS_DEV)
	docker build $(DIR)/docker/debian -f $(DIR)/docker/debian/Dockerfile.iptables -t $(BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES_DEV)
	docker build $(DIR)/docker/debian -f $(DIR)/docker/debian/Dockerfile.ssh -t $(BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH_DEV)

.PHONY: docker-push-dev-base-images
docker-push-dev-base-images:
	gcloud docker -- push $(BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES_DEV)
	gcloud docker -- push $(BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH_DEV)
	gcloud docker -- push $(BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS_DEV)

.PHONY: docker-build-and-push-dev-base-images
docker-build-and-push-dev-base-images: docker-build-base-images docker-push-dev-base-images

# local binaries
.PHONY: update-local-binary-cli
update-local-binary-cli:
	@bazel build //cmd/cli
	cp -f $(DIR)/bazel-bin/cmd/cli/cli $(DIR)/bin/lattice-system

# docker build hackery
.PHONY: docker-enter-build-shell
docker-enter-build-shell: docker-build-start-build-container
	docker exec -it $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh /bin/bash

.PHONY: docker-build-bazel-build
docker-build-bazel-build:
	docker build $(DIR)/docker -f $(DIR)/docker/bazel-builder/Dockerfile.bazel-build -t lattice-build/bazel-build

.PHONY: docker-build-start-build-container
docker-build-start-build-container: docker-build-bazel-build
	$(DIR)/docker/bazel-builder/start-build-container.sh

# cloud images
.PHONY: cloud-images-build
cloud-images-build: cloud-images-build-base-node-image cloud-images-build-master-node-image

.PHONY: cloud-images-build-base-node-image
cloud-images-build-base-node-image:
	$(CLOUD_IMAGE_BUILD_DIR)/build-base-node-image

.PHONY: cloud-images-build-master-node-image
cloud-images-build-master-node-image:
	$(CLOUD_IMAGE_BUILD_DIR)/build-master-node-image

.PHONY: cloud-images-clean
cloud-images-clean:
	rm -rf $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts

.PHONY: cloud-images-clean-master-node-image
cloud-images-clean-master-node-image:
	rm -rf $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts/master-node
	rm -f $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts/master-node-ami-id
