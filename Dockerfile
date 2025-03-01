FROM golang:1.23 AS build-stage
WORKDIR /app

RUN apt-get update && apt-get install -y \
    libssl-dev \
    libssl3

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY prism prism
RUN echo ls -la
RUN mkdir -p bin
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/backend -v ./prism/cmd/backend/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/worker -v ./prism/cmd/worker/main.go

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /app

COPY --from=build-stage /app/bin/* .

# Copy SSL libraries from their actual location
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libssl.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libcrypto.so* /lib/x86_64-linux-gnu/

# Copy other libraries
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libstdc++.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libgomp.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libgcc_s.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libc.so* /lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib64/ld-linux-x86-64.so* /lib64/