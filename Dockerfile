######################## builder stage ########################
FROM golang:1.24-alpine AS build
RUN apk add --no-cache make bash coreutils
RUN go install gotest.tools/gotestsum@latest
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

######################## runtime stage ########################
FROM alpine:latest AS runtime
COPY --from=build /src/coach_demon /coach_demon
COPY config.sample.yaml /config.yaml
ENTRYPOINT ["/coach_demon"]

######################## tests stage ########################
FROM build AS tests
