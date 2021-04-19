FROM golang:1.16 AS build
WORKDIR /app/src
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -x -o /example-db .

FROM alpine:latest
WORKDIR /root/
COPY --from=build /example-db .
EXPOSE 8080
CMD ["./example-db"]
