# Hakoniwa (箱庭)

<img src="https://raw.githubusercontent.com/aplulu/hakoniwa/main/ui/public/hakoniwa_logo.webp" width="200" alt="Hakoniwa Logo">

Hakoniwa is an On-Demand Desktop Environment service running on Kubernetes. It dynamically provisions lightweight desktop environments for users and serves them through a unified web interface.

## Features

*   **On-Demand Provisioning:** Automatically creates a Kubernetes Pod running a desktop environment (XFCE via Webtop) when a user logs in.
*   **Unified Gateway:** A custom Go proxy acts as the single entry point. It serves the React frontend for authentication/loading states and seamlessly switches to proxying traffic to the desktop instance once it's ready.
*   **Automatic Cleanup:** Background workers monitor inactivity and terminate unused instances to save resources.
*   **Kubernetes Native:** Fully integrated with Kubernetes for pod management using `client-go`.

## Architecture

*   **Frontend:** React (Vite + TypeScript + Radix UI Themes) - Handles authentication UI, status polling, and loading screens.
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
| `AUTH_AUTO_LOGIN` | If `true`, the frontend will automatically attempt to log in using the only available authentication method if there's just one configured in `AUTH_METHODS`. Default: `false`. | `false` |

#### OIDC Authentication Flow

When `oidc` is enabled, the authentication flow is as follows:
1.  User clicks the OIDC login button on the frontend.
2.  Frontend redirects the browser to `/_hakoniwa/api/auth/oidc/authorize`.
3.  The Hakoniwa backend generates an authorization URL and redirects the browser to the OIDC Identity Provider (IdP).
4.  User authenticates with the IdP.
5.  IdP redirects the browser back to the `OIDC_REDIRECT_URL` (which is a backend endpoint).
6.  The Hakoniwa backend processes the callback, exchanges the authorization code for tokens, sets a session cookie, and then redirects the browser to the frontend (`/_hakoniwa/`).
7.  If an error occurs during authentication, the backend redirects to the frontend with an error parameter (e.g., `/_hakoniwa/?error=login_failed`).

## License

MIT License
