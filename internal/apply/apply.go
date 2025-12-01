package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

// DryRunEnabled controls whether ApplyTarget performs a server-side dry-run
// before applying. Tests can set this to false to avoid fake-client dry-run issues.
var DryRunEnabled = true

func ApplyTarget(ctx context.Context, c client.Client, sourcePath string, target configsv1alpha1.TargetRef) error {
	logger := log.FromContext(ctx)

	// 1. List all YAML files
	files, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to list source directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !(strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		filePath := filepath.Join(sourcePath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// 2. Handle multi-document YAML
		docs := strings.Split(string(data), "\n---")
		for _, doc := range docs {
			doc = strings.TrimSpace(doc)
			if doc == "" {
				continue
			}

			// 3. Decode into Unstructured object
			obj := &unstructured.Unstructured{}
			jsonData, err := yaml.YAMLToJSON([]byte(doc))
			if err != nil {
				return fmt.Errorf("failed to convert YAML to JSON in %s: %w", filePath, err)
			}
			if err := obj.UnmarshalJSON(jsonData); err != nil {
				return fmt.Errorf("failed to unmarshal object in %s: %w", filePath, err)
			}

			// 4. Override namespace if object is namespaced
			if target.Namespace != "" {
				obj.SetNamespace(target.Namespace)
			}

			// 4.5 Clean metadata and status fields that must not be present for server-side apply/dry-run.
			cleanObjectForApply(obj)

			// 5. Dry-run validation: perform a server-side dry-run apply first to catch admission/validation errors
			applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("configsync")}
			dryRunOpts := append(applyOpts, client.DryRunAll)

			if DryRunEnabled {
				logger.Info("Performing dry-run apply")
				if err := c.Patch(ctx, obj, client.Apply, dryRunOpts...); err != nil {
					return fmt.Errorf("dry-run failed for %s from %s: %w",
						obj.GetKind(), filePath, err)
				}
			}

			// 6. Actual apply
			if err := c.Patch(ctx, obj, client.Apply, applyOpts...); err != nil {
				return fmt.Errorf("failed to apply %s from %s: %w",
					obj.GetKind(), filePath, err)
			}

			logger.Info("Applied manifest",
				"kind", obj.GetKind(),
				"name", obj.GetName(),
				"namespace", obj.GetNamespace(),
				"file", filePath,
			)
		}
	}

	return nil
}

// cleanObjectForApply removes fields that should not be sent to the API server
// when performing server-side apply or dry-run. In particular, some objects
// committed to repositories may include `metadata.managedFields` or other
// server-populated metadata which causes the API server to reject requests
// (e.g. "metadata.managedFields must be nil"). This function removes those
// fields in-place on the given Unstructured object.
func cleanObjectForApply(obj *unstructured.Unstructured) {
	if obj == nil {
		return
	}
	content := obj.UnstructuredContent()

	// Remove top-level status if present
	delete(content, "status")

	// Clean metadata map
	// remove known server-populated metadata fields and recursively strip any
	// nested managedFields that might appear in unexpected places.
	removeManagedFieldsRecursive(content)

	// write back content to object
	obj.SetUnstructuredContent(content)
}

// removeManagedFieldsRecursive walks an object represented as arbitrary
// interface{} (maps and slices) and removes keys named "managedFields",
// and also removes common server-populated metadata fields under any
// "metadata" map. This is defensive: some manifests may accidentally
// contain full kubectl output with managedFields embedded.
func removeManagedFieldsRecursive(v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		// If this map has metadata, clean known fields there.
		if metaI, ok := t["metadata"]; ok {
			if meta, ok := metaI.(map[string]interface{}); ok {
				delete(meta, "managedFields")
				delete(meta, "resourceVersion")
				delete(meta, "uid")
				delete(meta, "creationTimestamp")
				delete(meta, "generation")
				delete(meta, "selfLink")
				t["metadata"] = meta
			}
		}
		// Remove managedFields at this level if present
		delete(t, "managedFields")

		// Recurse into all values
		for _, vv := range t {
			removeManagedFieldsRecursive(vv)
		}
	case []interface{}:
		for _, e := range t {
			removeManagedFieldsRecursive(e)
		}
	default:
		// primitives: nothing to do
	}
}

// No rendering (not Helm)

// No Kustomize

// No drift detection (SSA handles merge)

// No pruning of deleted files yet (add later if desired)
