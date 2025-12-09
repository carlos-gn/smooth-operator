# How smooth-operator Works

## What is smooth-operator?

It's a **program that runs inside your Kubernetes cluster** and automates deploying MCP servers.

## The Big Picture:

```
┌─────────────────────────────────────────────────────────┐
│  Kubernetes Cluster                                     │
│                                                         │
│  ┌──────────────┐         ┌─────────────────┐          │
│  │  User        │         │ smooth-operator │          │
│  │  (You)       │         │   (watching)    │          │
│  └──────┬───────┘         └────────┬────────┘          │
│         │                          │                    │
│         │ 1. kubectl apply         │                    │
│         │    MCPServer YAML        │                    │
│         v                          │                    │
│  ┌─────────────────┐               │                    │
│  │  MCPServer      │───────────────┘                    │
│  │  Resource       │ 2. Operator sees it!               │
│  │  (Custom)       │                                    │
│  └─────────────────┘                                    │
│         │                          │                    │
│         │                          │ 3. Operator creates:
│         │                          v                    │
│         │                   ┌─────────────┐             │
│         │                   │   Pod       │             │
│         │                   │ (your MCP   │             │
│         │                   │  server)    │             │
│         │                   └─────────────┘             │
│         │                   ┌─────────────┐             │
│         │                   │  Service    │             │
│         │                   └─────────────┘             │
└─────────────────────────────────────────────────────────┘
```

## How it works:

### Step 1: You create an MCPServer resource

```yaml
apiVersion: mcp.dev/v1alpha1
kind: MCPServer
metadata:
  name: my-weather-server
spec:
  image: seville-weather:latest
  replicas: 2
  secretName: weather-secrets
```

### Step 2: smooth-operator is watching

The operator is a **Go program running in K8s** that:
- Watches for MCPServer resources (created, updated, deleted)
- When it sees one, it runs your **controller code**

### Step 3: Controller creates the actual resources

Your controller code (which we're about to write) does:
```go
// Pseudocode
func Reconcile(mcpserver) {
    // 1. Create a Deployment for the pods
    deployment := makeDeployment(mcpserver.Spec.Image, mcpserver.Spec.Replicas)
    kubernetes.Create(deployment)

    // 2. Create a Service to expose it
    service := makeService()
    kubernetes.Create(service)

    // 3. Update status
    mcpserver.Status.AvailableReplicas = 2
    mcpserver.Status.Phase = "Running"
}
```

## The smooth-operator directory structure:

```
smooth-operator/
├── api/v1alpha1/
│   └── mcpserver_types.go       ← YOU EDITED THIS
│                                  Defines MCPServer fields (image, replicas, etc.)
│
├── internal/controller/
│   └── mcpserver_controller.go  ← YOU WILL EDIT THIS NEXT
│                                  The code that creates pods/services
│
├── config/                      ← Generated K8s manifests
│   ├── crd/                      CRD definitions
│   ├── rbac/                     Permissions for the operator
│   └── manager/                  Deployment for the operator itself
│
├── Makefile                     ← Commands to build/deploy
└── main.go                      ← Entry point (starts the operator)
```

## The files you edited (mcpserver_types.go):

```go
type MCPServerSpec struct {
    Image      string  // User says: "run this container"
    Replicas   int32   // User says: "I want 2 copies"
    SecretName string  // User says: "use this secret for API keys"
}

type MCPServerStatus struct {
    AvailableReplicas int32  // Operator reports: "2 pods are ready"
    Phase             string // Operator reports: "Running"
}
```

**This is just the data structure.** It doesn't DO anything yet.

## The controller (mcpserver_controller.go) - what you'll write next:

This is the **brain** that does the work:

```go
func (r *MCPServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) {
    // 1. Get the MCPServer resource
    mcpserver := getMCPServer(req.Name)

    // 2. Create or update the Deployment
    deployment := &appsv1.Deployment{
        Spec: appsv1.DeploymentSpec{
            Replicas: &mcpserver.Spec.Replicas,
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{{
                        Image: mcpserver.Spec.Image,
                        // ... ports, env vars, etc.
                    }},
                },
            },
        },
    }
    createOrUpdate(deployment)

    // 3. Update the status
    mcpserver.Status.Phase = "Running"
    updateStatus(mcpserver)
}
```

**This Reconcile function runs EVERY time:**
- Someone creates an MCPServer
- Someone updates an MCPServer
- The pods change state
- Periodically (every few minutes)

## The reconciliation loop:

```
User creates MCPServer
        ↓
Controller sees it
        ↓
Controller creates Deployment
        ↓
Deployment creates Pods
        ↓
Controller updates MCPServer.Status
        ↓
[Wait for changes...]
        ↓
User updates MCPServer (changes replicas: 2 → 3)
        ↓
Controller sees change
        ↓
Controller updates Deployment
        ↓
New pod is created
        ↓
Controller updates MCPServer.Status
        ↓
[Loop continues...]
```

## What `make manifests` does:

Takes your Go code (`mcpserver_types.go`) and generates actual Kubernetes YAML files:

```yaml
# Generated in config/crd/bases/
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: mcpservers.mcp.dev
spec:
  group: mcp.dev
  names:
    kind: MCPServer
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          spec:
            properties:
              image:
                type: string
              replicas:
                type: integer
```

This teaches Kubernetes "MCPServer is a valid resource type now"

---

## Summary

The operator is:
1. **mcpserver_types.go** = defines what data MCPServer holds
2. **mcpserver_controller.go** = the code that creates pods when it sees MCPServer
3. **main.go** = starts the operator and watches for MCPServer resources

## What happens when you run the operator:

1. You deploy smooth-operator to your K8s cluster
2. It starts watching for MCPServer resources
3. You create an MCPServer YAML and apply it
4. The controller's Reconcile() function runs
5. It creates a Deployment with your specified image/replicas
6. K8s creates the pods
7. The controller updates the MCPServer status to show it's running
8. If you change the MCPServer, the controller sees it and updates the Deployment
9. The loop continues forever, keeping actual state matching desired state
