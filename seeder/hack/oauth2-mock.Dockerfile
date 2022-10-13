FROM node:16-alpine3.15

RUN npm install -g oauth2-mock-server

ENTRYPOINT [ "oauth2-mock-server", "-a 0.0.0.0", "-p 8080" ]
