# OpenShift Project Lister

A simple Go command-line tool that authenticates to an OpenShift cluster using username and password, then retrieves and displays a list of all projects (namespaces) the user has access to.

## Prerequisites

- Go 1.19 or later
- Access to an OpenShift cluster
- Valid username and password credentials

## Installation

1. Clone or navigate to this repository:
   ```bash
   cd /home/bryon/development/ocp-dev
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o ocp-lister ./cmd/ocp-lister
   ```

## Usage

Set the required environment variables and run the application:

```bash
export USER="myuser"
export PASSWORD="mypassword"
export SERVER="https://api.sno.bakerapps.net:6443"

./ocp-lister
```

Or run in a single command:

```bash
USER=myuser PASSWORD=mypassword SERVER=https://api.sno.bakerapps.net:6443 ./ocp-lister
```

## Environment Variables

- `USER` (required): OpenShift username
- `PASSWORD` (required): OpenShift password
- `SERVER` (required): OpenShift API server URL (e.g., `https://api.sno.bakerapps.net:6443`)

## Example Output

```
Connecting to OpenShift cluster at https://api.sno.bakerapps.net:6443...
Successfully authenticated!
Retrieving projects...

Found 5 project(s):

1. default
2. kube-public
3. kube-system
4. my-project
5. another-project
```

## Development

### Project Structure

```
ocp-dev/
├── cmd/
│   └── ocp-lister/
│       └── main.go          # Application entry point
├── internal/
│   ├── auth/
│   │   └── auth.go          # Authentication configuration
│   ├── client/
│   │   └── client.go        # Kubernetes client creation
│   └── projects/
│       └── projects.go      # Project listing logic
├── go.mod
├── go.sum
└── README.md
```

### Running During Development

```bash
go run ./cmd/ocp-lister
```

### Building

```bash
go build -o ocp-lister ./cmd/ocp-lister
```

## Security Notes

- Passwords are passed via environment variables for security
- Never commit credentials to version control
- Consider using a secrets manager for production use

## Troubleshooting

### Authentication Errors

If you see authentication errors:
- Verify your username and password are correct
- Ensure the SERVER URL is correct and accessible
- Check that your user has appropriate permissions

### Certificate Errors

If you encounter TLS/certificate errors, you may need to:
- Verify the cluster's CA certificate
- For development/testing with self-signed certs, you can modify `internal/client/client.go` to set `Insecure: true` (not recommended for production)

## License

[Add your license here]

