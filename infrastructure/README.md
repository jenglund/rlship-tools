# Infrastructure

This directory contains Kubernetes configurations and infrastructure-as-code for the Tribe application backend services.

## Directory Structure

```
infrastructure/
├── k8s/                    # Kubernetes configurations
│   ├── base/              # Base configurations
│   │   ├── namespace.yaml
│   │   └── secrets.yaml
│   ├── services/          # Service-specific configurations
│   │   ├── users/
│   │   ├── lists/
│   │   ├── activities/
│   │   └── interests/
│   ├── ingress/           # Ingress configurations
│   └── monitoring/        # Monitoring and observability
├── docker/                # Dockerfile for each service
└── scripts/              # Infrastructure management scripts
```

## Services

The following services are deployed:

- Users Service: User management and authentication
- Lists Service: Shared list management
- Activities Service: Activity tracking and management
- Interests Service: Interest button system and notifications

## Prerequisites

- Docker Desktop with Kubernetes enabled
- kubectl CLI tool
- helm (for installing ingress-nginx)

## Local Development Setup

1. Enable Kubernetes in Docker Desktop

2. Install ingress-nginx controller:
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx
```

3. Apply base configurations:
```bash
kubectl apply -f k8s/base/
```

4. Apply service configurations:
```bash
kubectl apply -f k8s/services/
```

5. Apply ingress configurations:
```bash
kubectl apply -f k8s/ingress/
```

## Monitoring

We use the following tools for monitoring and observability:
- Prometheus for metrics collection
- Grafana for visualization
- Jaeger for distributed tracing

## Development Workflow

1. Build service Docker images:
```bash
./scripts/build-images.sh
```

2. Deploy to local Kubernetes:
```bash
./scripts/deploy-local.sh
```

3. Access services:
- API Gateway: http://localhost/api
- Grafana: http://localhost/grafana
- Jaeger UI: http://localhost/jaeger

## Production Deployment

For production deployment, we use Google Kubernetes Engine (GKE). The deployment process is automated through GitHub Actions workflows. 