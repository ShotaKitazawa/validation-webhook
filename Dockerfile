## Build stage ##
FROM golang:1.13.3
WORKDIR /workdir
COPY . .
ARG IMAGE_TAG
ARG COMMIT_SHA
# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

## Run stage ##
FROM alpine:3.10.3
# Copy app binary
COPY --from=0 /workdir/app .
# Run
ENTRYPOINT ["./app"]
