package defaults

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var ReconcileFunc = reconcile.Func(func(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	fmt.Printf("event received for %s\n", request.NamespacedName)
	return reconcile.Result{}, nil
})
