# Deployment

## Local Development
```bash
docker-compose -f docker-compose.dev.yml up -d
make proto
make build
make test
```

## Kind Cluster
```bash
make kind-up
kubectl apply -f services/agent/manifests/
```

## Production (AWS)
```bash
cd infra/terraform/aws
terraform init
terraform apply
```

## Agent Installation
```bash
kr cluster add my-cluster
kubectl apply -f https://get.kuberoute.ai/agent.yaml
```

## CI/CD
GitHub Actions workflow:
1. Lint + test on PR
2. Build Docker images
3. Push to ECR
4. Deploy to staging (kind)
5. Integration tests
6. Deploy to production
