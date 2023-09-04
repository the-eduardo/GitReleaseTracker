# GitReleaseTracker

## Overview

GitReleaseTracker is a Discord bot designed to keep your community updated with the latest software releases from your GitHub repository. This bot sends messages to a Discord channel whenever new releases are published. Stay informed and make it easier for your users to access the latest updates.

## Table of Contents

- [Requirements](#Requirements)
- [Getting Started](#getting-started)
- [Important!](#Important!)
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

3. **Configuration:**
   - Add your Tokens `docker-compose.yml` file:

     ```env
     DISCORD_TOKEN=your-discord-bot-token
     GITHUB_TOKEN=your-github-api-token
     GITHUB_OWNER=github-account-owner # Who owns the repo
     GITHUB_REPO=github-repository # Repo name
     DISCORD_CHANNEL=discord-channel-id
     WAITING_TIME=60 # Minutes to wait before checking for new releases, max 1440 (24 hours)
     ```

4. **Run the Bot using docker compose:**

   ```bash
   docker-compose up -d --build
   ```
## Important!

My Dockerfile is set to build an application image that's compatible with `arm64` Linux architectures. This is determined by `GOARCH=arm64 GOOS=linux` in the Go build command within the Dockerfile.

If you need to run this on a different architecture, you'll need to modify these settings accordingly in the Dockerfile.

Here's a quick guide to adapting the Dockerfile for different environments:

- ### Changing the Architecture

The `GOARCH` environment variable sets the target architecture. You can replace `arm64` with your target architecture. Go supports several architectures like `amd64`, `386`, `arm`, `arm64` etc.

- ### Changing the Operating System

The `GOOS` environment variable sets the target operating system. You can replace `linux` with your target operating system. Go supports several operating systems like `windows`, `darwin` (for macOS), `linux`, etc.

Here's an example of how you might modify the Dockerfile to build for `amd64` architecture on a `windows` system:

```Dockerfile
# Stage 1: Build the Go application
FROM golang:1.21 AS builder
LABEL authors="the-eduardo"

WORKDIR /app
COPY main.go go.mod go.sum ./
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o app

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/app /

ENTRYPOINT ["/app"]
```

In this example, `GOARCH=amd64` and `GOOS=windows` target 64-bit Windows systems.

Ensure to check the compatibility of your target architecture and operating system with Go and your application.
## Understanding the Code
Written in Go and easily configurable using environment variables. Here are some key components:

- **Configuration (`readConfig` function):** Reads the values from environment variables and sets default values if not provided. It also performs basic validation of values.

- **Main Function:** Starts the Discord bot and GitHub client. It continually checks for new releases in the repository and sends messages to Discord when a new release is found.

## Contributing

Contributions to GitReleaseTracker are welcome! If you find a bug, have an enhancement in mind, or would like to propose a new feature, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
