GOARCH ?= $(shell go env GOARCH)
ifeq ($(GOARCH), arm)
DOCKER_ARG_ARCH=armv7
else
DOCKER_ARG_ARCH=$(GOARCH)
endif

DOCKER_IMAGE_NAME ?= hostport-allocator
DOCKER_REPO ?= local
DOCKER_IMAGE_TAG  ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

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

# Build manager binary
manager: fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -o bin/hostport-allocator-manager-linux-$(DOCKER_ARG_ARCH) main.go

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

# Create the kind cluster
kind:
	kind create cluster --name hostport-allocator --kubeconfig ./kind-kubeconfig

# Delete the kind cluster
kind-clean:
	kind delete cluster --name hostport-allocator

# Run tilt
tilt:
	KUBECONFIG=kind-kubeconfig tilt up --hud=true --no-browser

# Remove tilt
tilt-down:
	KUBECONFIG=kind-kubeconfig tilt down

# Build docker image
docker:
	docker build -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)-$(GOARCH)" --build-arg ARCH=$(DOCKER_ARG_ARCH) --build-arg OS="linux" .

# Tag docker image as latest
docker-latest:
	docker tag "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)-$(GOARCH)" "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):latest-$(GOARCH)"

# Push docker image
docker-push:
	docker push "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)-$(GOARCH)"

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
