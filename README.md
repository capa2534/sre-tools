# sre-tools

CLI tools built in Go for day-to-day SRE work. Built as a learning project to develop Go skills in a real Kubernetes/observability context.

---

## podcheck

List Kubernetes pods with status, restart count and CrashLoopBackOff detection.

### Usage
```bash
podcheck                    # all namespaces
podcheck -n monitoring      # filter by namespace
podcheck --only-errors      # show only problematic pods
podcheck --restarts 3       # show pods with 3+ restarts
```

### Example output
```
✗ crasher                    monitoring   CrashLoopBackOff ⚠️   restarts: 22
✓ loki-stack-0               monitoring   Running               restarts: 1
✓ coredns-76c974cb66-hqpbz   kube-system  Running               restarts: 3
```

---

## lokitail

Query Loki logs directly from the terminal without opening Grafana.

### Usage
```bash
lokitail -n monitoring                        # last 30m of logs
lokitail -n monitoring -since 1h             # last 1 hour
lokitail -n monitoring -filter "error"       # filter by text
lokitail -n monitoring -since 2h -filter "panic"
```

---

## grafana-summary

List configured Grafana alert rules via API.

### Usage
```bash
grafana-summary --token <service-account-token>
```

### Example output
```
Alertas configuradas: 1
⚠️  prometheus - High Restarts (for: 10m)
```

---

## upgrade-tracker

Kubernetes operator that tracks client upgrade status via a custom CRD.

### Usage
```bash
# Run the operator
upgrade-tracker

# List all upgrades
upgrade-tracker list
```

### Example output
```
CLIENT               VERSION         STATUS
client-abc           1.3.0           ✅ completed
client-def           1.5.0           ⏳ pending
client-xyz           1.4.0           🔄 in-progress
```

### Create an upgrade
```yaml
apiVersion: sre.io/v1
kind: ClientUpgrade
metadata:
  name: client-abc
  namespace: default
spec:
  clientName: client-abc
  targetVersion: "1.3.0"
  status: pending
```

---

## Stack

- Go 1.22+
- Kubernetes / k3s
- Grafana + Loki + Prometheus
- client-go, controller-runtime

## Learning context

Built with help of Claude IA and mix of 4 years of cloud experience - learning Go through real infrastructure problems instead of abstract tutorials
