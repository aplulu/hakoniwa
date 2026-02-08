# Hakoniwa (箱庭)

<img src="https://raw.githubusercontent.com/aplulu/hakoniwa/main/ui/public/img/hakoniwa_logo.webp" width="200" alt="Hakoniwa Logo">

Hakoniwa is an On-Demand Cloud Workspace Service running on Kubernetes. It dynamically provisions personalized cloud workspace environments (such as Webtop, Jupyter Notebook, or VS Code Server) for users and serves them through a unified web interface.

## Features

*   **On-Demand Provisioning:** Automatically creates a Kubernetes Pod for a selected workspace type when a user requests it.
*   **Multiple Instance Types:** Supports provisioning various workspace types (e.g., XFCE via Webtop, Jupyter Notebook, VS Code Server) from configurable Pod templates.
*   **Unified Gateway:** A custom Go proxy acts as the single entry point. It serves the React frontend for authentication/dashboard and seamlessly switches to proxying traffic to a user-selected workspace instance. Workspace selection is managed via a cookie, allowing the upstream application to receive requests at its root path (`/`).
*   **Multi-Instance Management:** Users can launch and manage multiple workspace instances of different types simultaneously.
*   **Automatic Cleanup:** Background workers monitor inactivity and terminate unused instances to save resources.
*   **Periodic Synchronization:** A background syncer periodically reconciles the in-memory instance list with the actual state of Pods in Kubernetes, ensuring consistency and handling external changes (e.g., manual Pod deletion).
*   **Kubernetes Native:** Fully integrated with Kubernetes for pod lifecycle management using `client-go`.
*   **User Management:** Supports anonymous and OIDC authentication.
*   **Instance Lifecycle:** Provides API and UI for creating, opening, and deleting workspace instances.

## Architecture

*   **Frontend:** React (Vite + TypeScript + Radix UI Themes) - Provides a dashboard for instance management, authentication UI, and loading screens.
*   **Backend:** Go (Standard Library + `client-go`) - Manages authentication, Kubernetes pod lifecycle, reverse proxying, and background synchronization/cleanup.
*   **Infrastructure:** Kubernetes - Orchestrates the application and user workspace instances.

## Getting Started

### Installation

Hakoniwa is designed to be deployed on Kubernetes using Helm.

1.  **Add the Helm Repository:**
    ```bash
    helm repo add hakoniwa https://aplulu.github.io/hakoniwa/
    helm repo update
    ```

2.  **Install the Chart:**
    ```bash
    helm install hakoniwa hakoniwa/hakoniwa \
      --set config.jwtSecret="<YOUR_SECRET_KEY>" \
      --namespace hakoniwa --create-namespace
    ```
    *Note: It is recommended to provide your own `config.jwtSecret`. If omitted, a random secret will be generated.*

## Configuration

Configuration is handled via environment variables:

| Variable | Description | Default |
| :--- | :--- |:-----------------------------|
| `LISTEN` | Address to listen on | `""` (All interfaces) |
| `PORT` | Port to listen on | `8080` |
| `KUBECONFIG` | Path to kubeconfig file (optional) | `""` |
| `KUBERNETES_NAMESPACE` | Namespace to manage pods in | `default` |
| `INSTANCE_INACTIVITY_TIMEOUT`| Duration before idle instances are reaped | `1m` |
| `MAX_POD_COUNT` | Maximum total concurrent pods (across all users) | `100` |
| `MAX_INSTANCES_PER_USER` | Maximum instances allowed per user | `5` |
| `MAX_INSTANCES_PER_USER_PER_TYPE` | Maximum instances of a specific type allowed per user | `3` |
| `ENABLE_PERSISTENCE` | Enable persistent storage feature globally | `true` |
| `POD_TEMPLATE_PATH` | Path to a Pod YAML template file. This file can contain multiple Pod definitions (as a Kubernetes List or multi-document YAML), where each `metadata.name` defines an instance type (e.g., "webtop", "jupyter"). | `""` (Uses embedded default) |
| `TITLE` | Application title | `Hakoniwa` |
| `MESSAGE` | Welcome message displayed below the title | `On-Demand Cloud Workspace Environment` |
| `LOGO_URL` | URL to the application logo | `/_hakoniwa/hakoniwa_logo.webp` |
| `TERMS_OF_SERVICE_URL` | URL to the terms of service | `""` |
| `PRIVACY_POLICY_URL` | URL to the privacy policy | `""` |

### Pod Template Configuration

The file specified by `POD_TEMPLATE_PATH` (or the embedded default) should be a Kubernetes Pod definition. For multiple instance types, it should contain a list of Pods (e.g., `kind: List` with `items`, or multiple YAML documents separated by `---`).

Each Pod definition **must** include:
*   `metadata.name`: This will be used as the unique ID for the instance type (e.g., "webtop", "jupyter").
*   `metadata.annotations`:
    *   `hakoniwa.aplulu.me/display-name`: (Optional) A human-readable name for the instance type, displayed in the UI. Defaults to `metadata.name` if not provided.
    *   `hakoniwa.aplulu.me/description`: (Optional) A description of the instance type, displayed in the UI.
    *   `hakoniwa.aplulu.me/image-url` or `hakoniwa.aplulu.me/logo-url`: (Optional) URL to an image/logo for the instance type.
    *   `hakoniwa.aplulu.me/port`: The target port of the application running in the Pod (e.g., "3000" for Webtop, "8888" for Jupyter). Defaults to "3000".
    *   `hakoniwa.aplulu.me/persistable`: (Optional) Set to `"true"` to allow OIDC authenticated users to enable persistent storage for this instance type.
    *   `hakoniwa.aplulu.me/volume-path`: (Optional) The path where the persistent volume will be mounted within the container. Defaults to `/config`.
    *   `hakoniwa.aplulu.me/volume-size`: (Optional) The size of the persistent volume claim to request. Defaults to `10Gi`.
    *   `hakoniwa.aplulu.me/volume-storage-class`: (Optional) The StorageClass to use for the persistent volume claim. If omitted, the cluster's default StorageClass is used.

Example for `pod_template.yaml`:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: webtop
  annotations:
    hakoniwa.aplulu.me/display-name: "Webtop (XFCE Desktop)"
    hakoniwa.aplulu.me/description: "Full-featured Linux desktop environment."
    hakoniwa.aplulu.me/image-url: "/_hakoniwa/img/webtop.avif"
    hakoniwa.aplulu.me/port: "3000"
    hakoniwa.aplulu.me/persistable: "true"
    hakoniwa.aplulu.me/volume-path: "/config"
spec:
  containers:
  - name: webtop
    image: lscr.io/linuxserver/webtop:latest
    env:
    - name: PUID
      value: "1000"
    - name: PGID
      value: "1000"
    # HAKONIWA_INSTANCE_ID and HAKONIWA_BASE_URL are injected by Hakoniwa automatically
    ports:
    - containerPort: 3000
---
apiVersion: v1
kind: Pod
metadata:
  name: jupyter
  annotations:
    hakoniwa.aplulu.me/display-name: "Jupyter Notebook"
    hakoniwa.aplulu.me/port: "8888"
spec:
  containers:
  - name: jupyter
    image: jupyter/base-notebook:latest
    env:
    - name: JUPYTER_ENABLE_LAB
      value: "yes"
    - name: NOTEBOOK_ARGS
      value: "--allow-root --ip=0.0.0.0 --port=8888 --no-browser --NotebookApp.base_url=$(HAKONIWA_BASE_URL)"
    ports:
    - containerPort: 8888
```

### Authentication Configuration

Hakoniwa supports multiple authentication methods which can be configured via environment variables.

| Variable | Description | Default |
| :--- | :--- | :--- |
| `AUTH_METHODS` | Comma-separated list of enabled authentication methods. Valid options: `anonymous`, `oidc`. | `anonymous` |
| `AUTH_AUTO_LOGIN` | If `true`, the frontend will automatically attempt to log in using the only available authentication method if there's just one configured in `AUTH_METHODS`. Default: `false`. | `false` |
| `OIDC_ISSUER_URL` | OpenID Connect Issuer URL (e.g., `https://accounts.google.com`). Required if `oidc` is in `AUTH_METHODS`. | `""` |
| `OIDC_CLIENT_ID` | OpenID Connect Client ID. Required if `oidc` is in `AUTH_METHODS`. | `""` |
| `OIDC_CLIENT_SECRET` | OpenID Connect Client Secret. Required if `oidc` is in `AUTH_METHODS`. | `""` |
| `JWT_SECRET` | Secret key used to sign the session JWTs. **Must be a strong, random string and kept confidential.** | `hakoniwa-secret-key` |
| `OIDC_REDIRECT_URL` | OpenID Connect Redirect URL. This should point to the backend callback endpoint. <br>Example: `https://<YourHostName>/_hakoniwa/api/auth/oidc/callback` | `""` |
| `OIDC_NAME` | Display name for the OIDC login button on the frontend. | `OpenID Connect` |
| `OIDC_SCOPES` | Comma-separated list of OIDC scopes to request. | `openid,profile` |
| `SESSION_EXPIRATION` | Duration for which the session JWT is valid. Accessing the service will extend the session if remaining time is less than half of this duration (sliding session). | `24h` |

#### OIDC Authentication Flow

When `oidc` is enabled, the authentication flow is as follows:
1.  User clicks the OIDC login button on the frontend.
2.  Frontend redirects the browser to `/_hakoniwa/api/auth/oidc/authorize`.
3.  The Hakoniwa backend generates an authorization URL and redirects the browser to the OIDC Identity Provider (IdP).
4.  User authenticates with the IdP.
5.  IdP redirects the browser back to the `OIDC_REDIRECT_URL` (which is a backend endpoint).
6.  The Hakoniwa backend processes the callback, exchanges the authorization code for tokens, sets a session cookie, and then redirects the browser to the frontend (`/_hakoniwa/`).
7.  If an error occurs during authentication, the backend redirects to the frontend with an error parameter (e.g., `/_hakoniwa/?error=login_failed`).

## Development

### Prerequisites

*   Go 1.25+
*   Node.js 22+
*   pnpm (Latest)
*   Docker
*   Kubernetes Cluster (local or remote)

### Development Environment

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
    pnpm install
    pnpm run dev
    ```



## License

MIT License