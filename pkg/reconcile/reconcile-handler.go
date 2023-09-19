package reconcile

import (
	"context"
	"github.com/SchweizerischeBundesbahnen/lot/pkg/lot-client"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler is a function type which performs specific logic within a reconcile.Reconciler reconcile Func.
type Handler func(ctx context.Context, object client.Object, cl lot_client.Client, scheme *runtime.Scheme) error

// HandlerFuncs is a struct which contains the only two types of supported handlers
type HandlerFuncs struct {
	CreateOrUpdateHandler Handler
	DeleteHandler         Handler
}
