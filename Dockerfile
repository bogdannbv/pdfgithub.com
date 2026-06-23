FROM golang:1.26-alpine AS build

WORKDIR /src

COPY . .

RUN apk add build-base \
    && go build -ldflags "-linkmode external -extldflags -static" -o pdfgh -a ./cmd

FROM alpine:3.24

WORKDIR /app

COPY ./static /app/static
COPY --from=build /src/pdfgh /app/pdfgh

CMD ["/app/pdfgh"]
