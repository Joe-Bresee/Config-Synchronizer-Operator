# TODO — Config Synchronizer Operator

Generated from `requirements.md` on 2025-11-25.

This file lists the ordered, actionable implementation steps to build the operator.

1. Initialize Kubebuilder Project
   - Create base Go/kubebuilder project.
   - Commands:
     - `kubebuilder init --domain example.io --repo github.com/<your-org>/config-synchronizer-operator`
   - Acceptance: `main.go`, `PROJECT`, `config/`, and `api/` exist.

2. Define CRD Types
   - Add `ConfigSync` API types in `api/v1alpha1/configsync_types.go` to match `requirements.md`.
   - Run: `make generate` and `make manifests`.
   - Acceptance: CRD YAML contains `ConfigSync` schema in `config/crd`.

3. Scaffold Controller
   - Generate controller scaffold: `kubebuilder create api --group configs --version v1alpha1 --kind ConfigSync`.
   - Ensure reconciler is registered in `main.go`.
   - Acceptance: controller compiles and is wired.

4. Implement Source Fetchers
   - Create `internal/source` package with `fetch_source(configsync) -> (map[string]any, error)`.
   - Implement Git fetcher using `go-git` and in-cluster fetchers for `ConfigMap` and `Secret` using controller-runtime client.

5. Implement Validation
   - Add `internal/validate` with `validate_config(data) -> error` to ensure YAML/JSON validity.

6. Implement Templating
   - Add `internal/template` with `render_template(data, target) -> data`.
   - Use Go `text/template` to start; optionally add Jinja2-style behavior later.

7. Apply Targets
   - Add `internal/target` with `apply_target(target, data)` to create/patch `ConfigMap`/`Secret` objects.
   - Respect `overwrite` semantics and preserve unmanaged keys when required.

8. Reconcile Loop & Watches
   - Implement the full reconcile flow: fetch, validate, render, apply, update status.
   - Add watches for referenced `ConfigMap`/`Secret` objects and implement periodic requeue using `refreshInterval`.

9. Status, Conditions & Events
   - Implement `.status` updates: `lastSyncedTime`, `sourceRevision`, `appliedTargets`, and `conditions`.
   - Emit Kubernetes events for success/failure.

10. RBAC, Manifests, Deployment
    - Add RBAC rules in `config/rbac` and operator `Deployment` manifest in `deploy/`.
    - Add `Dockerfile` and `Makefile` targets for building and pushing the image.

11. Testing & CI
    - Add unit tests for fetchers, validation, templating, and target apply logic.
    - Add integration tests with `envtest` and GitHub Actions for `go test`.

12. Docs & Examples
    - Add `README.md` usage examples and example `ConfigSync` manifests under `examples/`.

---

Usage
- Mark progress by editing this file or by using the repository's task tracking workflow.
- Recommended immediate actions: complete items 1–3 to scaffold the project, then implement 4 (Git fetcher) and 5 (validation).

If you'd like, I can scaffold parts of this (CRD types, controller skeleton, or `internal/source`), or follow whichever step you pick next.


<!-- idea: rollback support -->
<!-- multi-branch / env support -->
<!-- go back and fix kubebuilder validation for branch, revision and add branch to sync -->
<!-- //right now assuming https. Will need to add functionality for ssh later. Will need to make/reade secret for auth
Add full logging + error types + conditions updates

Generate unit tests for Git logic

Add compare-SHA logic in your Reconcile loop

Add server-side apply code for applying manifests -->

<!-- rbac
 -->
KNOWN HOST SUPPORT
gitignore for sensitive stuff when testing
- first probably raw manifest apply support - then include helm support.

richer error/fmt handling in fetch.go

Gaps & Risks (highest impact first)

Apply-loop incomplete: Without a robust apply_target, operator won't create/patch target ConfigMap/Secret as intended. Files: apply.go, reconciler apply loop.
Templating & validation missing: No template rendering or config validation; dangerous to apply raw source directly.
Condition handling bug: setCondition currently updates LastTransitionTime on every call — leads to noisy status churn. Needs logic to set LastTransitionTime only when the condition Status changes.
RBAC verification: Generated RBAC exists but verify it allows get/list/watch for Secrets/ConfigMaps and create/patch for applied resources. Also ensure the controller ServiceAccount is assigned required roles.
Tests coverage: No unit tests for fetchers/apply/templating. envtest/e2e not wired to validate full behavior.
Temp-dir management: Fetchers write to temp dirs; reconciler must defer os.RemoveAll or use an in-memory approach to avoid leaks.
Secrets handling: SSH/HTTPS credentials read from Secrets — ensure permissions and secure filesystem writes (mode 0600) are enforced (current code uses restrictive perms for secret files but confirm across all places).
Manifests generation: make manifests was previously flaky during earlier iterations — re-run to ensure CRD schemas reflect current api types.
Concrete next actions (prioritized)

Implement apply loop and target applier (highest priority)
Files to add/change: apply.go (or extend apply.go), update reconciler loop in configsync_controller.go to call applyTarget(ctx, r.Client, target, renderedData).
Acceptance: applying a sample ConfigSync causes target ConfigMap/Secret to be created/updated in cluster (can test on Kind).
Add templating and validation
Files: internal/template/render.go (use Go text/template or sprig functions), internal/validate/validate.go (YAML/JSON schema check).
Acceptance: renderer applies templates per Spec.Targets with test cases.
Fix setCondition behavior
Change: compute existing condition; only update LastTransitionTime when Status changes.
Files: modify configsync_controller.go setCondition.
RBAC & manifests verification
Run: make manifests, make generate, inspect config/rbac/* and config/crd/bases/*.
Ensure permissions for Secrets/ConfigMaps and server-side apply/patch are present.
Add unit tests for fetchers and apply
Files: internal/sources/*_test.go, internal/apply/*_test.go.
Use table tests and isolated temp dirs; for Git fetcher, use local git repo fixtures or go-git in-memory repos.
CI skeleton
Add GitHub Actions to run go test, go vet, make manifests on PRs.
Commands to run locally to validate current state

Build all packages: go build [repos](http://_vscodecontentref_/29).
Generate manifests (after code changes): make manifests
Run unit tests: go test [repos](http://_vscodecontentref_/30).
Build dev image: make dev-image (faster iteration)