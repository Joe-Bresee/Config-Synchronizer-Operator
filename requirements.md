# Config Synchronizer Operator — Copilot Requirements

## Project Summary
Build a Kubernetes operator that:
- Watches a `ConfigSync` Custom Resource (CR)
- Fetches configuration from a source (Git repo, ConfigMap, or Secret)
- Optionally applies templating or transformations
- Synchronizes it into one or more target ConfigMaps or Secrets across namespaces
- Updates `.status` with sync results

---

## CRD: ConfigSync

### spec
- `source` (one of)
  - `git`:
    - `repo` (string) — HTTPS/SSH URL
    - `path` (string) — path to file in repo
    - `revision` (string, optional, default `main`)
  - `configMapRef`: FIRST TODO
    - `name`
    - `namespace`
  - `secretRef`:
    - `name`
    - `namespace`
- `targets` (list)
  - `namespace`
  - `type` (enum: `ConfigMap` or `Secret`)
  - `name`
  - `key` (optional, for Secret)
- `refreshInterval` (string, optional)

### status
- `lastSyncedTime`
- `sourceRevision` (e.g., Git SHA)
- `appliedTargets` (int)
- `conditions` (list of conditions: Synced, Failed, InvalidSource)

---

## Operator Behavior

### Source Fetching
- TODO: `fetch_source(configsync: ConfigSync) -> dict`
- If Git: clone or pull, read file at `spec.source.git.path`
- If ConfigMap/Secret: read data from cluster
- Return dictionary of configuration

### Validation
- TODO: `validate_config(data: dict) -> bool`
- Ensure YAML/JSON is valid
- Optionally enforce schema rules
- Raise error or update `.status.conditions` on failure

### Templating / Transformation
- TODO: `render_template(data: dict, target: dict) -> dict`
- Apply simple variable interpolation
- Support Jinja2 (Python) or Go templates
- Optional: allow per-target overrides

### Target Application
- TODO: `apply_target(target: dict, data: dict)`
- Create or patch ConfigMap/Secret
- Preserve unmanaged keys unless `overwrite: true`
- Update `.status` after successful apply

### Reconciliation Triggers
- On CR create/update/delete
- On refresh interval
- On source changes (Git polling or ConfigMap/Secret watch)

### Error Handling
- Log errors with structured logging
- Update `.status.conditions` for Synced, Failed, InvalidSource
- Emit Kubernetes events
- Retry with exponential backoff

---

## Technical Stack

### Go (Kubebuilder)
- `controller-runtime`
- `go-git`
- `yaml.v3`
- Go templating package

---

## Directory Structure

config-synchronizer-operator/
README.md
requirements.md
outline.md
/api -> CRD types
/controllers -> reconciliation logic
/internal -> source fetch, parser, templates
/deploy -> manifests (CRD, RBAC, operator deployment)
Dockerfile
Makefile


---

## Example ConfigSync CR

apiVersion: configs.example.io/v1alpha1
kind: ConfigSync
metadata:
name: example-sync
spec:
source:
git:
repo: https://github.com/myorg/configs.git


path: config/app.yaml
revision: main
targets:
- namespace: default
type: ConfigMap
name: app-config
refreshInterval: 10m


---

## Stretch Goals
- Webhook triggers for Git push events
- SOPS/KMS encrypted secret support
- Multi-cluster sync
- Tekton pipeline config integration
- Jsonnet/Kustomize transformations

---

## Suggested TODOs for Copilot
- Create CRD YAML (`apiVersion: configs.example.io/v1alpha1`)
- Implement `fetch_source()` for Git and in-cluster objects
- Implement `validate_config()`
- Implement `render_template()`
- Implement `apply_target()`
- Implement main reconcile loop
- Implement status updates and condition reporting