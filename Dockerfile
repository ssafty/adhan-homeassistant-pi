FROM golang:1.20

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# slash at the end as mentioned in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Build
RUN go build -o /adhan-app

# Run
CMD [ "/adhan-app" ]
