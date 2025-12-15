# Smooth Operator

A Kubernetes operator for deploying and managing Model Context Protocol (MCP) servers.

*This is a learning project to understand Kubernetes operators and the controller pattern.*

## Overview

Smooth Operator automates the deployment of MCP servers in Kubernetes through custom resources. Define a simple `MCPServer` resource and the operator handles deployment, configuration, and lifecycle management.

## Features

- **Declarative deployment** - Define MCP servers as Kubernetes custom resources
- **Automatic reconciliation** - Operator maintains desired state
- **Secret management** - Automatic injection of environment variables from Kubernetes Secrets
- **Status reporting** - Track available replicas and deployment phase
- **Service creation** - Automatic ClusterIP service for stable pod access
- **Health checks** - Liveness and readiness probes for pod health monitoring
- **Observability** - Prometheus metrics and structured logging
- **Owner references** - Automatic resource cleanup when MCPServer is deleted
- **Configurable replicas** - Scale MCP server instances

## Architecture

```
User creates MCPServer resource
        ↓
Operator watches for changes
        ↓
Controller reconciles desired state
        ↓
Creates/updates Deployment & Service
        ↓
Kubernetes runs MCP server pods
        ↓
Status updated with available replicas & phase
        ↓
Metrics exposed for monitoring
```

**What the operator creates:**
- **Deployment** - Manages pod replicas with health checks
- **Service** - ClusterIP service for stable network access
- **Status updates** - Tracks deployment health and readiness

## Quick Start

### Prerequisites

- Go 1.21+
- kubectl
- kind (for local testing)
- Docker or Colima
- [Task](https://taskfile.dev/) (optional, Makefile also available)

### Installation

1. Create a local cluster:
   ```bash
   kind create cluster --name mcp-dev
   ```

2. Install CRDs:
   ```bash
   task install
   # or: make install
   ```

3. Run the operator:
   ```bash
   task run
   # or: make run
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
# Check MCPServer status
kubectl get mcpserver
kubectl describe mcpserver example-server

# Check created resources
kubectl get pods
kubectl get services
kubectl get deployments

# View operator logs
task run  # Shows structured logs with reconciliation events
```

### Monitoring

The operator exposes Prometheus metrics and health endpoints:

```bash
# View metrics (while operator is running)
curl http://localhost:8080/metrics

# Filter for MCPServer metrics
curl http://localhost:8080/metrics | grep mcpserver

# Health check
curl http://localhost:8081/healthz

# Readiness check
curl http://localhost:8081/readyz
```

**Available metrics:**
- `mcpserver_total` - Total number of MCPServer resources
- `mcpserver_phase` - MCPServers by phase (Running/Pending)
- `mcpserver_deployment_creation_errors_total` - Deployment creation failures
- `mcpserver_service_creation_errors_total` - Service creation failures
- `controller_runtime_reconcile_*` - Controller runtime metrics

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

### Available Tasks

```bash
# View all available tasks
task --list

# Common tasks
task manifests    # Generate CRD and RBAC manifests
task generate     # Generate DeepCopy methods
task test         # Run tests
task build        # Build operator binary
task dev          # Quick dev loop (install + run)
```

### Run tests

```bash
task test
# or: make test
```

### Build operator

```bash
task build
# or: make build
```

### Deploy to cluster

```bash
# Build and push image
task docker-build IMG=<registry>/smooth-operator:tag
task docker-push IMG=<registry>/smooth-operator:tag

# Deploy to cluster
task deploy IMG=<registry>/smooth-operator:tag

# Or use make
make docker-build docker-push deploy IMG=<registry>/smooth-operator:tag
```

### Development workflow

```bash
# After changing API types (api/v1alpha1/*_types.go)
task manifests    # Regenerate CRDs
task install      # Update CRDs in cluster
task run          # Restart operator

# After changing controller logic only
# Just restart operator (Ctrl+C then):
task run
```

### Testing

```bash
# Run integration tests
task test

# View coverage
go tool cover -html=cover.out

# Run tests with verbose output
go test ./... -v
```

**Test coverage:** 69.5% of controller code

Tests cover:
- Deployment creation with correct specs
- Service creation with correct specs
- Status updates
- Owner references
- Spec updates (replicas and image changes)

## Project Structure

```
├── api/v1alpha1/              # CRD schema definitions
│   └── mcpserver_types.go     # MCPServer spec and status
├── internal/controller/       # Reconciliation logic
│   ├── mcpserver_controller.go  # Main controller
│   └── metrics.go             # Prometheus metrics
├── config/                    # Kubernetes manifests
│   ├── crd/                   # Generated CRD definitions
│   ├── rbac/                  # Generated RBAC rules
│   ├── manager/               # Operator deployment
│   └── samples/               # Example MCPServer resources
├── cmd/main.go                # Operator entrypoint
├── Taskfile.yml               # Task automation
├── Makefile                   # Alternative build commands
└── LEARNING_NOTES.md          # Kubernetes operator concepts
```

## Roadmap

### Completed ✅
- [x] Status updates and health reporting
- [x] Service creation for pod exposure
- [x] Liveness and readiness probes (HTTP GET on /health)
- [x] Prometheus metrics and structured logging
- [x] RBAC configuration
- [x] Integration tests with 69.5% coverage
- [x] Conditional secret injection
- [x] Owner references for resource cleanup

### Planned 📋
- [ ] Helm chart for easy installation
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Deploy to real cluster
- [ ] Resource limit configuration
- [ ] Validation webhooks
- [ ] Multi-namespace support
- [ ] Custom printer columns for `kubectl get`

## License

Apache 2.0
