# k8s-autoscaler-cli

A simple CLI tool to monitor Kubernetes Deployment CPU/Memory usage and automatically scale replicas up or down based on thresholds.

## 📌 Features

- Fetch average CPU and Memory usage for a Kubernetes Deployment.
- Scale a deployment up or down by changing the replica count.
- Simple command-line interface using Cobra.
- Works with any K8s cluster you have access to via `kubectl`.

---

## 📦 Installation

### 1. Clone the repository

```bash
git clone https://github.com/tanbirali/k8s-autoscaler-cli.git
cd k8s-autoscaler-cli
```
