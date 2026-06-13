# Loadster

Cloud-Native Load Testing for Kubernetes and Modern Workloads

Loadster is an open-source, Kubernetes-native load testing platform built in Go. It is designed from the ground up to provide a frictionless developer experience with first-class support for Prometheus, Grafana, and modern DevOps workflows.

## Why Loadster?

Existing load testing solutions are powerful but often complex. They frequently require heavy runtime environments, complicated scripting, and external plugins to achieve basic observability. Loadster was built to solve these problems by prioritizing:

- **Developer First:** Run a load test locally or in a cluster in under 60 seconds with no complex configuration.
- **Batteries Included:** Built-in rich TUI dashboard and automated HTML/JSON report generation.
- **Observability First:** First-class Prometheus metrics integration out of the box.
- **Cloud Native:** Easily orchestrate distributed load generation across Kubernetes clusters using Helm and GitOps.

## Installation

You can install Loadster locally using Go or run it instantly via Docker. 

**Using Go:**
```bash
go install github.com/Shashank0701-byte/Loadster@latest
```

**Using Docker:**
```bash
docker pull ghcr.io/shashank0701-byte/loadster:latest
```

## Quick Start

Loadster makes it simple to scaffold, configure, and execute a load test.

### 1. Initialize a Scenario
Navigate to your working directory and run the initialization command. This will generate a default `test_scenario.yaml` file.

```bash
loadster init
```

### 2. Configure the Test
Edit the generated `test_scenario.yaml` to specify your target URL, headers, and load stages:

```yaml
target: https://api.example.com
headers:
  Content-Type: application/json
timeout: 5s
stages:
  - users: 10
    duration: 30s
  - users: 100
    duration: 1m
```

### 3. Run the Load Test
Execute the test. Loadster will automatically spin up a rich interactive terminal dashboard.

```bash
loadster run test_scenario.yaml
```

*Using Docker:*
```bash
docker run -it --rm -v ${PWD}:/data ghcr.io/shashank0701-byte/loadster run /data/test_scenario.yaml
```

## Distributed Execution

Loadster supports scaling traffic horizontally across a Kubernetes cluster. For detailed instructions on deploying Loadster workers via Helm, please refer to the documentation directory (coming soon).

## Project Status

Loadster is currently in active development. We welcome feedback, feature requests, and bug reports via GitHub Issues.

## License

This project is licensed under the MIT License - see the LICENSE file for details.