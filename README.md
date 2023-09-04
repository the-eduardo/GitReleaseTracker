# GitReleaseTracker

## Overview

GitReleaseTracker is a Discord bot designed to keep your community updated with the latest software releases from your GitHub repository. This bot sends messages to a Discord channel whenever new releases are published. Stay informed and make it easier for your users to access the latest updates.

## Table of Contents

- [Requirements](#Requirements)
- [Getting Started](#getting-started)
- [Note](#Note)
- [Understanding the Code](#understanding-the-code)
- [Contributing](#contributing)
- [License](#license)

## Requirements

Before you can run it, you need to have installed and configured:

- ### Docker

Docker is a platform that enables developers to build, share, and run applications in containers.
If you don't have Docker installed already, you can download it from the [official Docker website](https://docs.docker.com/get-docker/).

- ### Docker Compose

The easiest and recommended way to get Docker Compose is to install Docker Desktop. Docker Desktop includes Docker Compose along with Docker Engine and Docker CLI which are Compose prerequisites.
## Getting Started

To use GitReleaseTracker, follow these steps:

1. **Clone the Repository:**
```bash
git clone https://github.com/the-eduardo/GitReleaseTracker
```

2. **Configuration:**
   - Add your Tokens `docker-compose.yml` file:

     ```env
     DISCORD_TOKEN=your-discord-bot-token
     GITHUB_TOKEN=your-github-api-token
     GITHUB_OWNER=github-account-owner # Who owns the repo
     GITHUB_REPO=github-repository # Repo name
     DISCORD_CHANNEL=discord-channel-id
     WAITING_TIME=60 # Minutes to wait before checking for new releases, max 1440 (24 hours)
     ```

3. **Run the Bot using docker compose:**

   ```bash
   docker-compose up -d --build
   ```
## Note

The Dockerfile is currently configured to build an application image compatible with the `arm64` Linux architecture. This is specified in the Go build command within the Dockerfile using the `GOARCH=arm64` and `GOOS=linux` environment variables.

When your target environment differs, you **WILL** need to modify these settings in the Dockerfile. Here's a detailed guide on how to adapt the Dockerfile for different architectures and operating systems:

### Changing the Architecture

The `GOARCH` environment variable sets the target architecture for the build. If you are targeting a different architecture, replace `arm64` with your target architecture. Go supports multiple architectures, such as `amd64` for x86-64, `386` for x86, `arm` for 32-bit ARM, and `arm64` for 64-bit ARM, among others.

### Changing the Operating System

The `GOOS` environment variable sets the target operating system for the build. If you are targeting a different operating system, replace `linux` with your target OS. Go supports various operating systems, including `windows`, `darwin` (for macOS), `linux`, and more.

Here's an example of how to modify the Dockerfile to build for the `amd64` architecture on a Linux system:

```Dockerfile
# Stage 1: Build the Go application
FROM golang:1.21 AS builder
LABEL authors="the-eduardo"

WORKDIR /app
COPY main.go go.mod go.sum ./
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o app

# Stage 2: Create a minimal runtime image
FROM ubuntu:latest

# Install CA certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app /

ENTRYPOINT ["/app"]
```

In this example, `GOARCH=amd64` and `GOOS=linux` are set to build for 64-bit Linux systems.

Always verify the resulting Docker image. You can use the `docker run` command to run the image and check if it works as expected.
## Understanding the Code
Written in Go and easily configurable using environment variables. Here are some key components:

- **Configuration (`readConfig` function):** Reads the values from environment variables and sets default values if not provided. It also performs basic validation of values.

- **Main Function:** Starts the Discord bot and GitHub client. It continually checks for new releases in the repository and sends messages to Discord when a new release is found.

## Contributing

Contributions to GitReleaseTracker are welcome! If you find a bug, have an enhancement in mind, or would like to propose a new feature, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
