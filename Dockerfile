FROM golang:1.20

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# install mp3 related dependencies
RUN apt update && apt install -y libasound2-dev && rm -rf /var/lib/apt/lists/*

# Build
RUN go build -o /adhan-pi

# Run
CMD [ "/adhan-pi" ]
