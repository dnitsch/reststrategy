## Build
FROM golang:1.19-buster AS build

ARG VERSION=0.0.0

ARG REVISION=aaaa1111

WORKDIR /app

# Copy the Go Modules manifests
COPY ./ ./
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod vendor

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -mod=readonly -buildvcs=false \
-ldflags="-s -w -X \"github.com/dnitsch/reststrategy/kubebuilder-controller/cmd/controller.Version=${VERSION}\" -X \"github.com/dnitsch/reststrategy/kubebuilder-controller/cmd/controller.Revision=${REVISION}\" -extldflags -static" \
-o dist/controller ./cmd

RUN chmod +x ./dist/controller

## Deploy
FROM gcr.io/distroless/base-debian11

LABEL org.opencontainers.image.source=https://github.com/dnitsch/reststrategy/kubebuilder-controller
LABEL org.opencontainers.image.description="RestStrategy controller from kubebuilder"
LABEL org.opencontainers.image.licence=MIT

WORKDIR /

COPY --from=build /app/dist/controller controller

USER nonroot:nonroot

ENTRYPOINT ["/controller"]
