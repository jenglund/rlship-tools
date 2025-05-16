#!/bin/bash

# Exit on error
set -e

echo "Building Docker images..."
docker build -t tribe/api:latest -f infrastructure/docker/api.Dockerfile .
docker build -t tribe/web:latest -f infrastructure/docker/web.Dockerfile .

echo "Creating namespace..."
kubectl apply -f infrastructure/k8s/base/namespace.yaml

echo "Applying configurations..."
kubectl apply -f infrastructure/k8s/services/api/config.yaml

echo "Creating secrets..."
# Note: In production, secrets should be managed more securely
kubectl create secret generic api-secrets \
  --namespace=tribe \
  --from-literal=db_user=postgres \
  --from-literal=db_password=postgres \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Deploying services..."
kubectl apply -f infrastructure/k8s/services/api/deployment.yaml
kubectl apply -f infrastructure/k8s/services/api/service.yaml
kubectl apply -f infrastructure/k8s/services/web/deployment.yaml
kubectl apply -f infrastructure/k8s/services/web/service.yaml

echo "Configuring ingress..."
kubectl apply -f infrastructure/k8s/ingress/ingress.yaml

echo "Waiting for deployments to be ready..."
kubectl wait --namespace=tribe \
  --for=condition=ready pod \
  --selector=app=api \
  --timeout=90s
kubectl wait --namespace=tribe \
  --for=condition=ready pod \
  --selector=app=web \
  --timeout=90s

echo "Deployment complete! Services are accessible at:"
echo "Web Frontend: http://localhost"
echo "API: http://localhost/api" 