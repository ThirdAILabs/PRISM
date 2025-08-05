FROM golang:1.23 AS build-stage
WORKDIR /app

RUN apt-get update && apt-get install -y \
    libssl-dev \
    libssl3

RUN go install github.com/google/go-licenses@latest
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Install Playwright CLI globally for later use
RUN PWGO_VER=$(grep -oE "playwright-go v\S+" ./go.mod | sed 's/playwright-go //g') \
    && go install github.com/playwright-community/playwright-go/cmd/playwright@${PWGO_VER}

COPY COPYING .
COPY prism prism
RUN mkdir -p bin
# RUN CGO_ENABLED=1 GOOS=linux go build -o bin/backend -v ./prism/cmd/backend/main.go
# RUN CGO_ENABLED=1 GOOS=linux go build -o bin/worker -v ./prism/cmd/worker/main.go

RUN touch bin/backend
RUN touch bin/worker

RUN go-licenses save ./prism/cmd/backend --save_path ./third_party_licenses/backend || :
RUN go-licenses save ./prism/cmd/worker --save_path ./third_party_licenses/worker || :

# Change to Ubuntu for the final stage instead of distroless
FROM debian:bookworm-slim AS build-release-stage
WORKDIR /app

# Copy binaries
COPY --from=build-stage /app/bin/* ./
COPY --from=build-stage /go/bin/playwright /usr/local/bin/
COPY --from=build-stage /app/third_party_licenses ./third_party_licenses
COPY --from=build-stage /app/COPYING .

# Install Playwright dependencies
RUN apt-get update && apt-get install -y ca-certificates tzdata \
    && /usr/local/bin/playwright install --with-deps chromium \
    && rm -rf /var/lib/apt/lists/*

# Copy application data
COPY data data
COPY prism/services/resources resources

# Copy SSL libraries from their actual location
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libssl.so* /usr/lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libcrypto.so* /usr/lib/x86_64-linux-gnu/

# Copy other libraries
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libstdc++.so* /usr/lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libgomp.so* /usr/lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libgcc_s.so* /usr/lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib/x86_64-linux-gnu/libc.so* /usr/lib/x86_64-linux-gnu/
COPY --from=build-stage /usr/lib64/ld-linux-x86-64.so* /lib64/
