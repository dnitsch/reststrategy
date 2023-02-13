## Build
FROM golang:1.19-buster AS build

ARG VERSION=0.0.0

ARG REVISION=aaaa1111

WORKDIR /app

COPY ./ ./

RUN go mod vendor

RUN CGO_ENABLED=0 go build -mod=readonly \
-ldflags="-s -w -X \"github.com/dnitsch/reststrategy/controller/cmd/controller.Version=${VERSION}\" -X \"github.com/dnitsch/reststrategy/controller/cmd/controller.Revision=${REVISION}\" -extldflags -static" \
-o dist/controller ./cmd

RUN chmod +x ./dist/controller

## Deploy
FROM gcr.io/distroless/base-debian11

LABEL org.opencontainers.image.source=https://github.com/dnitsch/reststrategy/controller
LABEL org.opencontainers.image.description="RestStrategy controller"
LABEL org.opencontainers.image.licence=MIT

WORKDIR /

COPY --from=build /app/dist/controller controller

USER nonroot:nonroot

ENTRYPOINT ["/controller"]
