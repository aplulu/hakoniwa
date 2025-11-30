# Hakoniwa (箱庭)

<img src="https://raw.githubusercontent.com/aplulu/hakoniwa/main/ui/public/hakoniwa_logo.webp" width="200" alt="Hakoniwa Logo">

Hakoniwa is an On-Demand Desktop Environment service running on Kubernetes. It dynamically provisions lightweight desktop environments for users and serves them through a unified web interface.

## Features

*   **On-Demand Provisioning:** Automatically creates a Kubernetes Pod running a desktop environment (XFCE via Webtop) when a user logs in.
*   **Unified Gateway:** A custom Go proxy acts as the single entry point. It serves the React frontend for authentication/loading states and seamlessly switches to proxying traffic to the desktop instance once it's ready.
*   **Automatic Cleanup:** Background workers monitor inactivity and terminate unused instances to save resources.
*   **Kubernetes Native:** Fully integrated with Kubernetes for pod management using `client-go`.

## Architecture

*   **Frontend:** React (Vite + TypeScript + Tailwind CSS) - Handles authentication UI, status polling, and loading screens.
*   **Backend:** Go (Standard Library + `client-go`) - Manages authentication, Kubernetes pod lifecycle, and reverse proxying.
*   **Infrastructure:** Kubernetes - Orchestrates the application and user desktop instances.

## Getting Started

### Prerequisites

*   Go 1.25+
*   Node.js 22+
*   Docker
*   Kubernetes Cluster (local or remote)

### Local Development

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/aplulu/hakoniwa.git
    cd hakoniwa
    ```

2.  **Build and Run Locally:**
    You can use the provided Makefile to build the project.
    ```bash
    make build
    ./bin/hakoniwa
    ```
    *Note: For the backend to interact with a Kubernetes cluster locally, ensure your `KUBECONFIG` is set correctly or `~/.kube/config` is accessible.*

3.  **Run Frontend (Development Mode):**
    ```bash
    cd ui
    npm install
    npm run dev
    ```

### Deployment

Hakoniwa is designed to be deployed on Kubernetes using Kustomize.

1.  **Build the Docker Image:**
    ```bash
    make docker-build
    # Or for multi-platform
    make docker-buildx
    ```

2.  **Deploy to Kubernetes:**
    ```bash
    kubectl apply -k deployments/base
    ```

## Configuration

Configuration is handled via environment variables:

| Variable | Description | Default                      |
| :--- | :--- |:-----------------------------|
| `LISTEN` | Address to listen on | `""` (All interfaces)        |
| `PORT` | Port to listen on | `8080`                       |
| `KUBECONFIG` | Path to kubeconfig file (optional) | `""`                         |
| `KUBERNETES_NAMESPACE` | Namespace to manage pods in | `default`                    |
| `INSTANCE_INACTIVITY_TIMEOUT`| Duration before idle instances are reaped | `1m`                         |
| `MAX_POD_COUNT` | Maximum concurrent desktop instances | `3`                          |
| `POD_TEMPLATE_PATH` | Path to a custom Pod YAML template | `""` (Uses embedded default) |
| `TITLE` | Application title | `Hakoniwa` |
| `MESSAGE` | Welcome message displayed below the title | `On-Demand Cloud Desktop Environment` |
| `LOGO_URL` | URL to the application logo | `/_hakoniwa/hakoniwa_logo.webp` |
| `TERMS_OF_SERVICE_URL` | URL to the terms of service | `""` |
| `PRIVACY_POLICY_URL` | URL to the privacy policy | `""` |

## License

MIT License
