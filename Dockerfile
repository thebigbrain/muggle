FROM golang:alpine

WORKDIR /go/src/muggle

EXPOSE 80

ENV PORT=80

COPY ./vendor ./vendor
COPY go.mod .
COPY go.sum .
COPY ./httpserver/ .

RUN go build -mod=vendor -tags=jsoniter -o main
RUN chmod +x ./main

CMD ["./main"]