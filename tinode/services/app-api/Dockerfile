FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/app-api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /out/app-api /usr/local/bin/app-api
USER 65532:65532
EXPOSE 8080 8081
ENTRYPOINT ["/usr/local/bin/app-api"]

