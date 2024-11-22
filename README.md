# Ubuntu Server Setup CLI

A cli program written in Go to help setup Ubuntu servers for php applications.

## Usage

Build the binary for linux with `env GOOS=linux go build -o bin/setup`
Use SCP to put the binary file on the new server and run the executable as the root user.

## Test Environment

Build the binary for Linux with `env GOOS=linux go build -o bin/setup`

Then build the docker container and run it to test the cli in a fresh ubuntu environment.

`docker build -t ubuntu-sandbox .`

`docker run -ti ubuntu-sandbox /bin/bash`

Then run `setup` in the container to test thes CLI.

## Known Issues

* UFW does not work inside of the Docker container causing the script to exit with an error

* Options where user input is required does not behave properly.

* Need to implement dynamic inputs for git url, project name, etc.

* Need to generate a SSH key and output it to the user so they can import the key into GitHub.