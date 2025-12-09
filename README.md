# Smooth Operator

A Kubernetes operator for deploying and managing Model Context Protocol (MCP) servers.

*This is a learning project to understand Kubernetes operators and the controller pattern.*

## Overview

Smooth Operator automates the deployment of MCP servers in Kubernetes through custom resources. Define a simple `MCPServer` resource and the operator handles deployment, configuration, and lifecycle management.

## Features

- **Declarative deployment** - Define MCP servers as Kubernetes custom resources
- **Automatic reconciliation** - Operator maintains desired state
- **Secret management** - Automatic injection of environment variables from Kubernetes Secrets
- **Owner references** - Resource cleanup when MCPServer is deleted
- **Configurable replicas** - Scale MCP server instances

## Architecture

```
User creates MCPServer resource
        ↓
Operator watches for changes
        ↓
Controller reconciles desired state
        ↓
Creates/updates Deployment
        ↓
Kubernetes runs MCP server pods
```

## Quick Start

### Prerequisites

- Go 1.21+
- kubectl
- kind (for local testing)
- Docker

### Installation

1. Create a local cluster:
   ```bash
   kind create cluster
   ```

2. Install CRDs:
   ```bash
   make install
   ```

3. Run the operator:
   ```bash
   make run
   ```

### Usage

Create an MCPServer resource:

```yaml
apiVersion: mcp.mcp.dev/v1alpha1
kind: MCPServer
metadata:
  name: example-server
spec:
  image: my-mcp-server:latest
  replicas: 2
  port: 8080
  secretName: mcp-secrets
```

Apply it:

```bash
kubectl apply -f mcpserver.yaml
```

Watch it deploy:

```bash
kubectl get mcpserver
kubectl get pods
```

## API Reference

### MCPServer Spec

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | Yes | - | Container image for the MCP server |
| `replicas` | int32 | No | 1 | Number of pod replicas |
| `port` | int32 | No | 8080 | HTTP port the server listens on |
| `secretName` | string | No | - | Name of Secret containing environment variables |

### MCPServer Status

| Field | Type | Description |
|-------|------|-------------|
| `availableReplicas` | int32 | Number of ready pods |
| `phase` | string | Current deployment phase |
| `conditions` | []Condition | Standard Kubernetes conditions |

## Development

### Run tests

```bash
make test
```

### Build operator

```bash
make build
```

### Deploy to cluster

```bash
make docker-build docker-push IMG=<registry>/smooth-operator:tag
make deploy IMG=<registry>/smooth-operator:tag
```

## Project Structure

```
├── api/v1alpha1/              # CRD schema definitions
├── internal/controller/       # Reconciliation logic
├── config/                    # Kubernetes manifests
│   ├── crd/                   # CRD definitions
│   ├── rbac/                  # RBAC rules
│   └── samples/               # Example resources
└── Makefile                   # Build commands
```

## Roadmap

- [ ] Status updates and health reporting
- [ ] Service creation for pod exposure
- [ ] Liveness and readiness probes
- [ ] Resource limit configuration
- [ ] Prometheus metrics
- [ ] Validation webhooks
- [ ] Multi-namespace support

## License

Apache 2.0
