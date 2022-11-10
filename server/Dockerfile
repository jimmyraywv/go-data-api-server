FROM golang:1.19.2-alpine3.16 as builder
RUN apk add git build-base
WORKDIR /
COPY . ./
# Disable default GOPROXY
RUN go env -w GOPROXY=direct
RUN GOOS=linux GOARCH=amd64 go build -a -o main.bin .

# FROM scratch
FROM alpine:3.16
WORKDIR /
COPY --from=builder main.bin main
EXPOSE 8080
ENTRYPOINT ["/main"]
