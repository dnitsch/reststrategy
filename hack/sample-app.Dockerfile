FROM ghcr.io/muchobien/pocketbase:latest

RUN mkdir /pb_data /pb_public

EXPOSE 8090

ENTRYPOINT ["/usr/local/bin/pocketbase", "serve", "--http=0.0.0.0:8090", "--dir=/pb_data", "--publicDir=/pb_public"]