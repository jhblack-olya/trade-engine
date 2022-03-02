# FROM golang:alpine as builder
# RUN apk update && apk add --no-cache git

# WORKDIR /app
# # Copy go mod and sum files
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN go build -o main .

# #############################
# FROM alpine:latest
# RUN apk --no-cache add ca-certificates

# WORKDIR /root/
# COPY --from=builder /app/main .
# COPY --from=builder /app/conf.json .
# # COPY --from=builder /app/.env .  NO FILE AT the moment

# EXPOSE 8080
# CMD ["./main"]

############################# from line 1-22, its the proper way to run go, but for some reason, its not working.

FROM golang:alpine
RUN apk update && apk add --no-cache git
RUN cat /etc/resolv.conf
WORKDIR /app
COPY . .
RUN go mod download
RUN ls -al
RUN pwd
ENTRYPOINT ["go","run","main.go"]
EXPOSE 8080
# # CMD ["./main"]
