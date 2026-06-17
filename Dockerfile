FROM golang:1.26-alpine AS build

WORKDIR /src

COPY . .

RUN apk add build-base \
    && go build -ldflags "-linkmode external -extldflags -static" -o pdfgh -a .

FROM alpine:3.24

COPY ./static /app/static
COPY --from=build /src/pdfgithub /app/pdfgh

EXPOSE 80

CMD ["/app/pdfgh"]
