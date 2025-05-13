## Boxy — a Minimal Container CLI on top of containerd + runc

> **Boxy** is a single-binary command-line tool that gives you the familiar
> `pull / run / stop / rm / ps` workflow while re-using the proven containerd
> engine under the hood.

---

### 🗂️ Features

| Capability       | Notes                                                                                   |
| ---------------- | --------------------------------------------------------------------------------------- |
| `pull`           | Downloads + unpacks an OCI image (progress spinner).                                    |
| `run`            | Interactive shell by default; `-d` flag for detached mode. Auto-pulls image if missing. |
| `stop [timeout]` | Sends `SIGTERM`, waits *timeout* (default 10 s), escalates to `SIGKILL`.                |
| `rm [-f]`        | Deletes container + snapshot. `-f` kills first if still running.                        |
| `ps`             | Shows **NAME / STATE / PID / IMAGE** for all containers in the `minidock` namespace.    |

---

### 🛠️ Installation

#### Prerequisites

* Linux kernel ≥ 5.4
* `containerd` v2.x running as a system service
* Go 1.22+ (only for building)

#### Build & install from source

```bash
git clone https://github.com/<you>/boxy.git
cd boxy
go build -o boxy ./cmd/boxy        # creates ./boxy
mkdir -p $HOME/.local/bin
mv boxy $HOME/.local/bin/
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

*(system-wide: `sudo mv boxy /usr/local/bin/`)*

#### One-line install (after you push the repo)

```bash
go install github.com/<you>/boxy/cmd/boxy@latest   # binary lands in $HOME/go/bin
```

---

### 🚀 Quick start

```bash
boxy pull alpine            # download alpine:latest
boxy run --name demo alpine # interactive shell
/ # exit
boxy run -d --name web alpine sleep 300   # detached
boxy ps                   # list containers
boxy stop web             # graceful stop
boxy rm web               # delete
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
<summary><code>boxy run --name &lt;id&gt; [-d] &lt;image&gt; [cmd...]</code></summary>

* Interactive (default) uses the image’s default CMD or your override.
* Detached `-d` runs in background with no TTY.

```bash
boxy run --name api alpine        # /bin/sh
boxy run -d --name redis redis:7  # background
```

</details>

<details>
<summary><code>boxy ps</code></summary>

Shows running/stopped containers.

```
NAME   STATE     PID   IMAGE
api    RUNNING   2419  docker.io/library/alpine:latest
redis  STOPPED   -     docker.io/library/redis:7
```

</details>

<details>
<summary><code>boxy stop &lt;name&gt; [timeout]</code></summary>

Graceful shutdown.

```bash
boxy stop redis 5s
```

</details>

<details>
<summary><code>boxy rm [-f] &lt;name&gt;</code></summary>

Remove container and snapshot. Use `-f` to kill first.

```bash
boxy rm api
boxy rm -f redis
```

</details>

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
```

* Networking, volumes, and build support are pluggable and can be added later
  (CNI, BuildKit, etc.).

---

### 🌱 Roadmap ideas

| Priority | Planned feature                                             |
| -------- | ----------------------------------------------------------- |
| ⭐⭐⭐      | `logs <name>` (stream stdout/stderr of detached containers) |
| ⭐⭐       | `-p HOST:CONT` via CNI bridge + portmap                     |
| ⭐        | BuildKit integration (`boxy build -t myapp .`)              |
| ⭐        | Push / login to a local registry (`registry:2` or ORAS)     |

---

### 🤝 Contributing

1. Fork the repo & create a feature branch.
2. Follow *golangci-lint run* (no warnings).
3. Make PRs small and focused.

---

### 📜 License

MIT © 2025 – free to use, modify, and distribute.
