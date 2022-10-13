## Build
FROM golang:1.19-buster AS build

ARG VERSION=0.0.0

ARG REVISION=aaaa1111

WORKDIR /app

COPY ./ ./

RUN make VERSION=${VERSION} REVISION=${REVISION} build

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /app/dist/controller controller

RUN chmod +x /controller

USER nonroot:nonroot

CMD [ "/controller", "version" ]

ENTRYPOINT ["/controller"]
