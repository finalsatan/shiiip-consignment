FROM golang:alpine as builder

RUN apk --no-cache add git

WORKDIR /app/shiiip-consignment

COPY . .

RUN GOPROXY=https://goproxy.io go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shiiip-consignment


FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir /app
WORKDIR /app

COPY --from=builder /app/shiiip-consignment/shiiip-consignment .

CMD ["./shiiip-consignment"]