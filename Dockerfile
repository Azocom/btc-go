FROM golang:1.10

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

RUN go mod tidy

# This container exposes port 8080 to the outside world
# EXPOSE 8080
ENV PORT=8080

# Run the executable
CMD ["/go/src/app/src/btcgo"]

# https://codefresh.io/docs/docs/example-catalog/ci-examples/golang-hello-world/
#gcloud run deploy --source . apiitaguai --region=southamerica-east1