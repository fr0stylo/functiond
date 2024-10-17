# Functiond

Functiond is a lightweight, container-based function execution system built with Go and ContainerD. It provides a
scalable way to run serverless functions in isolated containers with automatic scaling and lifecycle management.

## Features

- Container-based function isolation using ContainerD
- Automatic scaling of function instances
- Network isolation per function
- ZIP-based function deployment
- Efficient snapshot management for quick container startup
- HTTP API for function execution
- Automatic instance cleanup and resource management

## Prerequisites

- Go 1.20 or later
- ContainerD v2.x
- Linux environment (for ContainerD support)
- Docker registry access (for base images)
- ContainerD CNI setup

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/functiond.git

# Build the project
cd functiond
go build
```

## Configuration

The system uses the following default configuration:

- ContainerD socket: `/run/containerd/containerd.sock`
- Default namespace: `example`
- Function directory: `/opt/function`
- Default base image: `docker.io/library/node:lts-alpine`
- HTTP server port: `:8080`

## Usage

### Preparing a Function

1. Create your function code
2. Package it in a ZIP file with the required dependencies
3. Place the ZIP file in an accessible location

### Starting the Server

```bash
sudo ./functiond
```

### Executing a Function

Send a GET request to execute your function:

```bash
curl http://localhost:8080/execute
```

## Architecture

### Components

1. **WorkerSet**: Manages a pool of function instances
    - Handles scaling and lifecycle management
    - Maintains worker pool based on demand
    - Implements automatic cleanup of idle instances

2. **Worker**: Individual function instance
    - Runs in isolated container
    - Handles function execution
    - Manages network isolation

3. **Snapshotter**: Manages container images
    - Creates and maintains container snapshots
    - Handles function code deployment
    - Optimizes container startup time

4. **Manager**: Orchestrates multiple WorkerSets
    - Registers and deregisters function versions
    - Manages function lifecycles
    - Handles resource allocation

## Development

### Project Structure

```
functiond/
├── main.go           # Server entry point
├── pkg/
│   ├── manager.go    # WorkerSet management
│   ├── runner/
│   │   ├── worker_set.go    # Worker pool management
│   │   ├── worker/
│   │   │   ├── worker.go    # Container instance management
│   │   └── snapshotter/
│   │       └── snapshotter.go # Image and code management
```

### Key Features Implementation

#### Auto-scaling

- Automatically scales workers based on demand
- Implements downscale timeout for resource efficiency
- Maintains worker pool within concurrency limits

#### Container Management

- Uses ContainerD for container lifecycle management
- Implements efficient snapshot-based deployments
- Handles cleanup of terminated containers

#### Network Isolation

- Provides network isolation per function instance
- Manages network attachment and detachment
- Supports custom network configurations

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- ContainerD team for the container runtime
- Go team for the programming language
