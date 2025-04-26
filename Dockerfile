######################## builder stage ########################
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

######################## runtime stage ########################
FROM alpine:latest AS runtime
COPY --from=build /src/coach_demon /coach
COPY config.sample.yaml /config.yaml
ENTRYPOINT ["/coach"]

######################## tests stage ########################
FROM build AS tests
# Install make inside the tests stage
RUN apk add --no-cache make bash coreutils
RUN go install gotest.tools/gotestsum@latest
# nothing else needed: we have /src with full code
