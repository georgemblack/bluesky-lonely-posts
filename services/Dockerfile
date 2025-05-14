FROM golang:1.24 AS build

WORKDIR /app
COPY . .

ENV GOOS=linux
ENV GOARCH=arm64
ENV CGO_ENABLED=0
RUN go build -o intake cmd/intake/main.go
RUN go build -o server cmd/server/main.go

FROM alpine

COPY --from=build /app/intake /intake
COPY --from=build /app/server /server

CMD ["/intake"]
