# Build the operator binary
FROM --platform=$BUILDPLATFORM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.25.1-alpine3.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workdir
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY components/common components/common
COPY components/operator components/operator

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} GOFIPS140=v1.0.0 go build -a -o operator ./components/operator/cmd/main.go

FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.22.1 AS certs

RUN apk add --no-cache ca-certificates

# Replace main images in values.yaml
# Bumping registry-proxy images in the config/registry-proxy/values.yaml - used for building "local" and "dev" images
FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.22.1 AS replacer

WORKDIR /workdir

ARG PURPOSE
ARG IMG_DIRECTORY
ARG IMG_VERSION
ARG CONTAINER_REGISTRY

ENV IMG_DIRECTORY=$IMG_DIRECTORY
ENV IMG_VERSION=$IMG_VERSION
ENV PROJECT_ROOT="."
ENV CONTAINER_REGISTRY=$CONTAINER_REGISTRY

RUN apk update && apk add --no-cache bash yq

COPY config/registry-proxy config/registry-proxy
COPY hack/ hack/

RUN if [[ "dev" = "$PURPOSE" ]] ; then ./hack/replace_registry_proxy_chart_images.sh all ; fi
RUN if [[ "local" = "$PURPOSE" ]] ; then ./hack/replace_registry_proxy_chart_images.sh main-only ; fi
# do nothing (keep unchanged versions) for "release"


# Use distroless as minimal base image to package the operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#TODO: check if we need gcr.io/distroless/static:nonroot here
FROM scratch

ENV GODEBUG=fips140=only,tlsmlkem=0

WORKDIR /

COPY --from=builder /workdir/operator .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chown=1000:1000 --from=replacer /workdir/config/registry-proxy /module-chart
USER 1000:1000

ENTRYPOINT ["/operator"]
