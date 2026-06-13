# Kubernetes & Observability

Loadster is built to be a first-class citizen in a Kubernetes ecosystem. You can deploy it using our Helm chart to easily run distributed load tests from within your cluster, avoiding local network bottlenecks.

## Deploying with Helm

You can install Loadster into your Kubernetes cluster using the provided Helm chart.

### 1. Prerequisites
Ensure you have `helm` and `kubectl` configured to point to your cluster.

### 2. Installation
Navigate to the root of the Loadster repository and run:

```bash
helm upgrade --install loadster ./deploy/helm/loadster \
  --namespace loadster-testing \
  --create-namespace
```

### 3. Configuring the Distributed Workers
By default, the Helm chart spins up Loadster in "distributed" mode. You can scale the number of load generation workers by updating the `replicaCount` in your `values.yaml` or via the CLI:

```bash
helm upgrade --install loadster ./deploy/helm/loadster \
  --namespace loadster-testing \
  --set replicaCount=5
```

## Observability (Prometheus & Grafana)

Loadster exposes metrics natively. It does not require sidecars or complex scraping proxies.

### Prometheus Integration
When deployed via Helm, Loadster automatically deploys a `ServiceMonitor` (if the Prometheus Operator is installed in your cluster). This tells Prometheus to automatically discover and scrape metrics from Loadster's `/metrics` endpoint.

**Available Metrics:**
- `loadster_requests_total`: Total number of requests sent
- `loadster_requests_errors`: Total number of failed requests
- `loadster_latency_ms`: Histogram of request latencies
- `loadster_active_users`: Number of currently active virtual users

### Grafana Dashboards
You can import our official Grafana dashboards (coming soon to the `dashboards/` directory) to visualize:
1. Target Success Rates
2. P50, P95, and P99 Request Latency
3. Network throughput and virtual user scale over time

If you are running locally without Kubernetes, you can easily verify metrics by navigating to `http://localhost:8080/metrics` while a test is actively running.
