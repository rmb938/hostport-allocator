DOCKER_ARCHS ?= amd64 armv7 arm64
BUILD_DOCKER_ARCHS = $(addprefix docker-,$(DOCKER_ARCHS))
PUSH_DOCKER_ARCHS = $(addprefix docker-push-,$(DOCKER_ARCHS))
LATEST_DOCKER_ARCHS = $(addprefix docker-latest-,$(DOCKER_ARCHS))

BUILD_GO_ARCHS = $(addprefix manager-,$(DOCKER_ARCHS))

DOCKER_IMAGE_NAME ?= hostport-allocator
DOCKER_REPO ?= local
DOCKER_IMAGE_TAG  ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

CRD_OPTIONS ?= "crd:crdVersions=v1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

.PHONY: manager $(BUILD_GO_ARCHS)
manager: fmt vet $(BUILD_GO_ARCHS)
$(BUILD_GO_ARCHS): manager-%:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(if $(filter $*,armv7),arm,$*) go build -o bin/hostport-allocator-manager-linux-$* main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

helm-crds:
	mkdir -p deploy/charts/hostport-allocator/templates/crds
	go run ./tools/crd-helm/main.go

# Create the kind cluster
kind:
	kind create cluster --name hostport-allocator --kubeconfig ./kind-kubeconfig

# Delete the kind cluster
kind-clean:
	kind delete cluster --name hostport-allocator

# Run tilt
tilt:
	KUBECONFIG=kind-kubeconfig tilt up --legacy=true

# Remove tilt
tilt-down:
	KUBECONFIG=kind-kubeconfig tilt down

.PHONY: docker $(BUILD_DOCKER_ARCHS)
docker: $(BUILD_DOCKER_ARCHS)
$(BUILD_DOCKER_ARCHS): docker-%:
	docker build -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-$*:$(DOCKER_IMAGE_TAG)" \
		--build-arg ARCH="$*" \
		--build-arg OS="linux" \
		.

.PHONY: docker-latest $(LATEST_DOCKER_ARCHS)
docker-latest: $(LATEST_DOCKER_ARCHS)
$(LATEST_DOCKER_ARCHS): docker-latest-%:
	docker tag "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-$*:$(DOCKER_IMAGE_TAG)" "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-$*:latest"

.PHONY: docker-push $(PUSH_DOCKER_ARCHS)
docker-push: $(PUSH_DOCKER_ARCHS)
$(PUSH_DOCKER_ARCHS): docker-push-%:
	docker push "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-$*:$(DOCKER_IMAGE_TAG)"

.PHONY: docker-manifest
docker-manifest:
	DOCKER_CLI_EXPERIMENTAL=enabled docker manifest create -a "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" $(foreach ARCH,$(DOCKER_ARCHS),$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-$(ARCH):$(DOCKER_IMAGE_TAG))
	$(foreach ARCH,$(DOCKER_ARCHS),DOCKER_CLI_EXPERIMENTAL=enabled docker manifest annotate --os linux --arch $(if $(filter $(ARCH),armv7),arm,$(ARCH)) $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-$(ARCH):$(DOCKER_IMAGE_TAG);)
	DOCKER_CLI_EXPERIMENTAL=enabled docker manifest push "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
KIND = $(LOCALBIN)/kind

## Tool Versions
KUSTOMIZE_VERSION ?= v4.4.0
CONTROLLER_TOOLS_VERSION ?= v0.8.0
KIND_VERSION ?= v0.11.1

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v4@v4.5.4

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
