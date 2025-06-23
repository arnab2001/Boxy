# boxy

A tiny container CLI powered by **containerd** and **runc**.

🔥 **No Docker daemon needed** — just containerd (system service) + Linux namespaces.

---

## 🚀 Quick start

```bash
# Install containerd + runc (Ubuntu/Debian example)
sudo apt update && sudo apt install -y containerd runc

# Clone & build boxy
git clone https://github.com/arnab2001/boxy.git
cd boxy && go build -o boxy ./cmd/boxy

# Pull an image
sudo ./boxy pull nginx

# Run a container with port forwarding
sudo ./boxy run --name web -p 8080:80 nginx

# Access your container
curl http://localhost:8080

# List containers
sudo ./boxy ps

# Stop and remove
sudo ./boxy stop web
sudo ./boxy rm web
```

---

### 📑 Command reference

<details>
<summary><code>boxy pull &lt;image&gt;</code></summary>

Download & unpack an image into containerd.

```bash
boxy pull nginx:1.27
```

</details>

<details>
<summary><code>boxy run --name &lt;id&gt; [-d] [-p HOST:CONT] &lt;image&gt; [cmd...]</code></summary>

* Interactive (default) uses the image's default CMD or your override.
* Detached `-d` runs in background with no TTY.
* Port forwarding `-p HOST:CONT[/PROTOCOL]` maps host ports to container ports.

```bash
# Basic container
boxy run --name api alpine        # /bin/sh

# Background container
boxy run -d --name redis redis:7  # background

# Port forwarding examples
boxy run --name web -p 8080:80 nginx                    # TCP (default)
boxy run --name app -p 3000:3000/tcp -p 5353:53/udp app # Multiple ports
boxy run --name db -p 127.0.0.1:5432:5432 postgres     # Bind to specific IP
```

**Port Publishing Syntax:**
- `-p 8080:80` - Map host port 8080 to container port 80 (TCP)
- `-p 8080:80/tcp` - Explicit TCP protocol
- `-p 9000:9000/udp` - UDP protocol
- `-p 127.0.0.1:5432:5432` - Bind to specific host IP (coming soon)

</details>

<details>
<summary><code>boxy ps</code></summary>

Shows running/stopped containers.

```
NAME   STATE     PID   IMAGE
web    RUNNING   2419  docker.io/library/nginx:latest
redis  STOPPED   -     docker.io/library/redis:7
```

</details>

<details>
<summary><code>boxy stop &lt;name&gt; [timeout]</code></summary>

Graceful shutdown with automatic network cleanup.

```bash
boxy stop redis 5s
```

</details>

<details>
<summary><code>boxy rm [-f] &lt;name&gt;</code></summary>

Remove container and snapshot with network cleanup.

```bash
boxy rm web
boxy rm -f redis  # force kill first
```

</details>

---

### 🌐 Networking & Port Publishing

Boxy uses **CNI (Container Network Interface)** for networking with automatic port forwarding via iptables.

#### Requirements
- CNI plugins installed at `/opt/cni/bin/` (bridge, portmap)
- iptables for port forwarding rules

#### Install CNI Plugins
```bash
# Download and install CNI plugins
wget https://github.com/containernetworking/plugins/releases/download/v1.4.1/cni-plugins-linux-amd64-v1.4.1.tgz
sudo mkdir -p /opt/cni/bin
sudo tar -xzf cni-plugins-linux-amd64-v1.4.1.tgz -C /opt/cni/bin
```

#### Network Configuration
Boxy automatically creates a bridge network (`boxy0`) with:
- **Root mode**: `172.18.0.0/16` subnet
- **Rootless mode**: `10.88.0.0/16` subnet

#### Rootless Support
Run boxy without root privileges:

```bash
# Rootless mode (experimental)
./boxy run --name app -p 8080:80 nginx
```

**Rootless Limitations:**
- Privileged ports (<1024) require `bypass4netns` plugin
- Some network features may be limited
- User namespace restrictions apply

---

### 🔍 Architecture snapshot

```
boxy CLI
   │ gRPC
   ▼
containerd (system daemon)
   │ fork/exec
   ▼
runc  →  Linux namespaces, cgroups
   │
   ▼
CNI plugins → bridge + iptables (port forwarding)
```

---

### 🌱 Roadmap ideas

| Priority | Status | Planned feature                                             |
| -------- | ------ | ----------------------------------------------------------- |
| ⭐⭐⭐      | ✅     | `-p HOST:CONT` via CNI bridge + portmap                     |
| ⭐⭐⭐      | 🔄     | `logs <name>` (stream stdout/stderr of detached containers) |
| ⭐⭐       | 📋     | BuildKit integration (`boxy build -t myapp .`)              |
| ⭐        | 📋     | Push / login to a local registry (`registry:2` or ORAS)     |
| ⭐        | 📋     | Volume mounts and bind mounts                               |

**Legend:** ✅ Complete | 🔄 In Progress | 📋 Planned

---

### 🤝 Contributing

1. Fork the repo & create a feature branch.
2. Follow *golangci-lint run* (no warnings).
3. Make PRs small and focused.
4. Add tests for new functionality in the `test/` directory.

---

### 🧪 Testing

```bash
# Run all tests
cd test && go test -v .

# Run benchmarks
go test -bench=.

# Run with coverage
go test -cover .
```

---

### 📝 License

MIT License - see [LICENSE](LICENSE) for details.
