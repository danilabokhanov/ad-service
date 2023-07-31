# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Download and install any required dependencies
RUN go mod download

# Build the Go app
RUN go build cmd/main/main.go

# Expose port 18080 for incoming traffic
EXPOSE 18080

# Define the command to run the app when the container starts
CMD ["/app/main"]
