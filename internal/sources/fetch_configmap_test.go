package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFetchConfigMap_WritesFilesAndReturnsSHA(t *testing.T) {
	ctx := context.Background()

	// create a fake client with a ConfigMap
	cm := &corev1.ConfigMap{}
	cm.Namespace = "myns"
	cm.Name = "mycm"
	cm.Data = map[string]string{"a.txt": "hello", "b.txt": "world"}

	sch := runtime.NewScheme()
	if err := corev1.AddToScheme(sch); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	fakeClient := fake.NewClientBuilder().WithScheme(sch).WithObjects(cm).Build()

	ref := &configsv1alpha1.ObjectRef{Namespace: "myns", Name: "mycm"}

	sha, dir, err := fetchConfigMap(ctx, fakeClient, ref)
	if err != nil {
		t.Fatalf("fetchConfigMap failed: %v", err)
	}
	defer os.RemoveAll(dir)

	// ensure files exist
	for filename, expected := range cm.Data {
		p := filepath.Join(dir, filename)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("failed to read file %s: %v", p, err)
		}
		if string(b) != expected {
			t.Fatalf("unexpected content in %s: %s", p, string(b))
		}
	}

	if len(sha) != 64 {
		t.Fatalf("expected sha length 64, got %d", len(sha))
	}
}
