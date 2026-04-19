package main

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var upgradeGVK = schema.GroupVersionKind{
	Group:   "sre.io",
	Version: "v1",
	Kind:    "ClientUpgrade",
}

type ClientUpgradeReconciler struct {
	client.Client
}

func (r *ClientUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(upgradeGVK)

	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	spec := obj.Object["spec"].(map[string]interface{})
	clientName := spec["clientName"].(string)
	targetVersion := spec["targetVersion"].(string)
	status := spec["status"].(string)

	fmt.Printf("🔄 Reconciling: %s | version: %s | status: %s\n", clientName, targetVersion, status)

	return ctrl.Result{}, nil
}

func (r *ClientUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(upgradeGVK)
	return ctrl.NewControllerManagedBy(mgr).
		For(obj).
		Complete(r)
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	scheme := runtime.NewScheme()
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Group: "sre.io", Version: "v1"})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		fmt.Printf("Error creando manager: %v\n", err)
		os.Exit(1)
	}

	if err := (&ClientUpgradeReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		fmt.Printf("Error configurando controller: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "list" {
		listUpgrades()
		return
	}
	fmt.Println("🚀 Upgrade Tracker corriendo...")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		fmt.Printf("Error iniciando manager: %v\n", err)
		os.Exit(1)
	}
}

