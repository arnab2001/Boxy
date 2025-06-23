# Boxy: Docker-style `-p HOST:CONT` Port Publishing Implementation Plan

## 1. Dependencies & Setup
- Add `github.com/containernetworking/cni/pkg/client` and related CNI plugin types to `go.mod`.
- Ensure `bridge`, `portmap`, and (optionally) `bypass4netns` binaries are installed at `/opt/cni/bin/`.
- Ship `/etc/cni/net.d/10-boxy-bridge.json` as described, or generate it if missing.

## 2. CLI & Flag Parsing
- Extend `run` command:
  - Add `--publish/-p` flag:
    ```go
    var portFlags []string
    cmd.Flags().StringSliceVarP(&portFlags, "publish", "p", nil, "HOST:CONT[,PROTO]")
    ```
  - Parse into a slice of CNI port mappings:
    - Implement `parsePorts(flags []string) ([]cnitypes.PortMapping, error)`
    - Validate and support both `tcp` and `udp` (default to `tcp`).

## 3. CNI Integration
- After `task.Start()` in `run.go`:
  - Get the container's netns:
    ```go
    netnsPath := fmt.Sprintf("/proc/%d/ns/net", task.Pid())
    ```
  - Call CNI to attach the namespace and set up port mapping:
    ```go
    result, err := cniConfig.AddNetwork(
        ctx, "boxy-bridge",
        cni.WithNamespacePath(netnsPath),
        cni.WithInterfaceName("eth0"),
        cni.WithPortMapping(portMappings),
    )
    ```
  - Store the result for possible inspection/diagnostics.

## 4. Cleanup on Stop/Rm
- In `stop.go` and `rm.go`:
  - Before/after killing the task, call:
    ```go
    _ = cniConfig.DelNetwork(ctx, "boxy-bridge",
            cni.WithNamespacePath(netnsPath))
    ```
  - Handle `errdefs.IsNotFound` gracefully.

## 5. Rootless Support
- Detect rootless mode (`os.Geteuid() != 0`):
  - If rootless, ensure `bypass4netns` is in the CNI config/plugins.
  - Document that `iptables-nftables-compat` must be in `PATH` for userns.
- Test port binding <1024: Use `bypass4netns` for unprivileged port exposure.

## 6. Testing & Validation
- Test Matrix:
  - Single and multiple containers with different port mappings.
  - Stop/rm removes DNAT rules.
  - Rootless mode with and without privileged ports.
- Automate tests (if possible) in CI.

## 7. Documentation
- Update README:
  - Usage examples for `-p`.
  - Rootless caveats and requirements.
  - How to install CNI plugins and config.

## 8. (Optional) Future Enhancements
- Compose-like `--network` options.
- Dynamic CNI config generation.
- IPv6 and service discovery.

## Suggested Implementation Order
1. Add flag parsing & port struct (½ day)
2. Integrate CNI add/del logic (1 day)
3. Update stop/rm for cleanup (½ day)
4. Rootless support & docs (1 day)
5. Test matrix & CI (½ day) 