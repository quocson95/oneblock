# Use the official Golang image to create a build artifact.
# This image has the Go compiler pre-installed.
FROM golang:1.23.4 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .
ENV GOCACHE=/root/.cache/go-build
# Build the Go app
RUN --mount=type=cache,target="/root/.cache/go-build" go build -ldflags="-s -w" -o oneblock_backend .
RUN chmod +x oneblock_backend

# Start a new stage from scratch
FROM golang:1.23.4
# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/oneblock_backend /app/oneblock_backend
COPY chart /app/chart
COPY raw_data /app/raw_data
COPY static /app/static
COPY config.json /app/config.json
EXPOSE 8080
# Expose port 8080 to the outside world
# Command to run the executable
CMD ["/app/oneblock_backend"]
