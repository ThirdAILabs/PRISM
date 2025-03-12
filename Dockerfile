FROM golang:1.23 AS build-stage
WORKDIR /app

RUN apt-get update && apt-get install -y \
    libssl-dev \
    libssl3

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Install Playwright CLI globally for later use
RUN PWGO_VER=$(grep -oE "playwright-go v\S+" ./go.mod | sed 's/playwright-go //g') \
    && go install github.com/playwright-community/playwright-go/cmd/playwright@${PWGO_VER}

COPY prism prism
RUN mkdir -p bin
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/backend -v ./prism/cmd/backend/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/worker -v ./prism/cmd/worker/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/train_worker_ndbs -v ./prism/cmd/train_worker_ndbs/main.go

# Change to Ubuntu for the final stage instead of distroless
FROM debian:bookworm-slim AS build-release-stage
WORKDIR /app

# Copy binaries
COPY --from=build-stage /app/bin/* ./
COPY --from=build-stage /go/bin/playwright /usr/local/bin/

# Install Playwright dependencies
RUN apt-get update && apt-get install -y ca-certificates tzdata \
    && /usr/local/bin/playwright install --with-deps firefox \
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

RUN WORK_DIR=/app/.worker_work_dir \
    UNIVERSITY_DATA=/app/data/university_webpages.json \
    DOC_DATA=/app/data/docs_and_press_releases.json \
    AUX_DATA=/app/data/auxiliary_webpages.json \
    PRISM_LICENSE=413E68-DC19F1-C0FFDB-1A7273-4F66CE-V3 \
    ./train_worker_ndbs