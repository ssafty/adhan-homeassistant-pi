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

# Copy the sample adhan.mp3 file. 
COPY adhan.mp3 ./

# install mp3 related dependencies
RUN apt update && apt install -y libasound2-dev && rm -rf /var/lib/apt/lists/*

# Build
RUN go build -o /adhan-homeassistant-pi

# Create a custom "exec mode" docker entrypoint to receive SIGTERM.
# `docker-entrypoint.sh` is not populated in the docker file because it
# requires --build-arg (build time variable replacement). Adding it to
# the repository (i.e. ENV runtime) avoids the frustration from a 2 steps 
# variables debugging process.
COPY docker-entrypoint.sh ./
RUN chmod 755 docker-entrypoint.sh 
ENTRYPOINT [ "./docker-entrypoint.sh"]