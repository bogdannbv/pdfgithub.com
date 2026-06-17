FROM golang:1.26-alpine AS build

WORKDIR /src

COPY . .

RUN apk add build-base \
    && go build -ldflags "-linkmode external -extldflags -static" -o pdfgithub -a .

FROM scratch

COPY ./static /static
COPY --from=build /src/pdfgithub /pdfgithub

EXPOSE 80

CMD ["/pdfgithub"]
