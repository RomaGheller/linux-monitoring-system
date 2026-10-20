# Linux Monitoring System

A lightweight Linux service monitoring project written in **Go**. It provides a long-running monitoring daemon and a separate command-line client for managing monitored services. The project is being developed as part of a Master's project focused on monitoring and automated recovery of Linux infrastructure on Raspberry Pi.

> **Project status:** Monitoring functionality is implemented and has been tested on Linux and Raspberry Pi. Automated recovery and the new deployment subsystem are planned, **not yet implemented** in this repository.

## Features implemented

- **Monitoring daemon** (`cmd/monitoring`) that periodically checks registered services.
- **CLI client** (`cmd/monitoringctl`) for listing and registering services.
- **systemd health checker** that reads service states via `systemctl`.
- **Unix domain socket API** for communication between the CLI and daemon.
- **Service registry** backed by YAML configuration.
- **Modular Go packages** separating monitoring logic, API, application orchestration, persistence, and systemd integration.
- **Automated Go tests** for major components, including API communication and monitoring logic.
- **Graceful shutdown** and context-based lifecycle management.

## Architecture

```text
                    Linux host / Raspberry Pi

     User / automation
             |
             v
     +-----------------+
     |  monitoringctl  |  CLI
     +--------+--------+
              |
              | Unix domain socket
              v
     +-----------------+
     | monitoring      |  daemon
     | application/API |
     +--------+--------+
              |
       +------+-------+
       |              |
       v              v
  Monitoring       Service registry
  engine           (YAML configuration)
       |
       v
  systemd checker
       |
       v
  systemctl / Linux services
```

### Repository layout

| Path | Responsibility |
| --- | --- |
| `cmd/monitoring` | Daemon entry point |
| `cmd/monitoringctl` | CLI entry point |
| `internal/monitoring` | Monitoring engine, checkers, scheduling, result handling |
| `internal/api` | Unix socket server/client and request/response handling |
| `internal/application` | Application-level coordination and checker creation |
| `internal/registry` | Service registration and persistence |
| `internal/config` | Configuration loading and validation |
| `internal/systemd` | systemd checker and command execution abstractions |
| `internal/testutil` | Test doubles and supporting utilities |

The code is organized under `internal/` to keep implementation packages private to this Go module.

## Requirements

- Linux with **systemd** (for systemd service checks)
- Go toolchain compatible with the version declared in [`go.mod`](go.mod)

The daemon uses a local Unix socket (in the current deployment, `/run/monitoring/monitoring.sock`). Creating and accessing this socket requires appropriate directory permissions. **Do not run an unreviewed daemon with root privileges.**

## Build and test

From the repository root:

```bash
go test ./...

go build -o bin/monitoring ./cmd/monitoring
go build -o bin/monitoringctl ./cmd/monitoringctl
```

The `bin/` directory is generated locally and is not committed to Git.

## Configuration and usage

The service registry uses a YAML file (`services.yaml`). A minimal empty configuration is:

```yaml
services: []
```

The following CLI commands illustrate a typical workflow **after the daemon has been configured and started**:

```bash
./bin/monitoringctl services list
./bin/monitoringctl services ensure ssh --type systemd
./bin/monitoringctl services list
```

`ssh` is an example of a systemd unit name; the actual unit name varies across distributions (for example, `sshd`). The daemon and CLI must have permission to access the configured socket and service registry. See the source in `cmd/` for current startup options; deployment setup is environment-specific.

## CI/CD and deployment context

The project is developed and deployed in a Raspberry Pi environment using a self-hosted **Forgejo** instance and **Forgejo Actions**. The broader infrastructure has demonstrated Go builds, publishing versioned artifacts, and deployment experiments. Infrastructure workflows and privileged deployment helpers are maintained separately; this repository is focused on the monitoring application.

## Roadmap

The following components are **planned / under design**, not implemented monitoring features:

- **Artifact storage:** a Go service for validating and storing deployable artifacts.
- **Service launcher:** a separate Go service for controlled deployment and lifecycle operations through systemd.
- **Automated recovery:** restore a previously known-good service version after an unhealthy state is detected.
- Additional checkers (for example, HTTP and container-based services).
- Persistent monitoring history and richer health reporting.

## Why Go?

Go supports a small deployable service with explicit error handling, straightforward concurrency (`goroutines`, `context`), interface-based separation of responsibilities, and a standard toolchain for testing and cross-compilation.

## Author

Roman Gheller — Master's project / independent Go development.
