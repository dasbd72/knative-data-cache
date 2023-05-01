FROM golang:1.20 AS build-stage

WORKDIR /app

# modules
COPY go.mod go.sum ./
RUN go mod download

# source code
COPY cmd cmd
COPY internal internal
# build
RUN CGO_ENABLED=0 GOOS=linux go build -o /image-scale cmd/image-scale/image-scale.go

# Deploy the application binary into a lean image
FROM alpine:edge AS release-stage

WORKDIR /

COPY --from=build-stage /image-scale /image-scale

EXPOSE 9090

ENTRYPOINT ["/image-scale"]