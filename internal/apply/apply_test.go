package apply

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApplyTarget_AppliesConfigMap(t *testing.T) {
	ctx := context.Background()

	// disable dry-run for fake client
	DryRunEnabled = false

	// prepare a temporary directory with a simple ConfigMap manifest
	tmpDir, err := os.MkdirTemp("", "apply-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmYAML := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
data:
  key: value
`

	filePath := filepath.Join(tmpDir, "cm.yaml")
	if err := os.WriteFile(filePath, []byte(cmYAML), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// build fake client with no existing objects
	sch := runtime.NewScheme()
	if err := corev1.AddToScheme(sch); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	fakeClient := fake.NewClientBuilder().WithScheme(sch).Build()

	target := configsv1alpha1.TargetRef{Namespace: "default"}

	if err := ApplyTarget(ctx, fakeClient, tmpDir, target); err != nil {
		t.Fatalf("ApplyTarget failed: %v", err)
	}

	// verify ConfigMap exists in fake client
	var cm corev1.ConfigMap
	if err := fakeClient.Get(ctx, client.ObjectKey{Namespace: "default", Name: "test-cm"}, &cm); err != nil {
		t.Fatalf("failed to get applied ConfigMap: %v", err)
	}

	if val, ok := cm.Data["key"]; !ok || val != "value" {
		t.Fatalf("unexpected data in ConfigMap: %v", cm.Data)
	}
}
