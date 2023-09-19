package reconcile

import (
	"context"
	lot_client "github.com/SchweizerischeBundesbahnen/lot/pkg/lot-client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler is nests the reconciler.Reconciler interface of the controller-runtime library. It is used in order to hide
// the underlaying library within the operator package
type Reconciler interface {
	reconcile.Reconciler
}

// WithClient is a function that returns a Reconciler with an opinionated reconcile method which can pass to the event
// handler functions not only its context but also a client.Client and the runtime.Scheme of the Operator's manager.Manager
func WithClient(cl lot_client.Client, obj client.Object, scheme *runtime.Scheme, fn *HandlerFuncs) Reconciler {
	return reconcile.Func(func(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
		var o client.Object
		switch reflect.TypeOf(obj).String() {
		case "*unstructured.Unstructured":
			o = copyUntypedObject(obj, scheme)
		default:
			o = copyTypedObject(obj)
		}

		log := logf.Log.WithName("ReconcileFunc")
		ctx = logf.IntoContext(ctx, log)
		log.V(1).Info("event received for", "namespaceName", request.NamespacedName, "kind", o.GetObjectKind())

		err := cl.Get(ctx, request.NamespacedName, o)
		if err != nil {
			log.Info("object not found", "resource", request.NamespacedName)
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}

		if fn.DeleteHandler != nil {
			if err := fn.DeleteHandler(ctx, o, cl, scheme); err != nil {
				return reconcile.Result{}, err
			}
		}

		if fn.CreateOrUpdateHandler != nil {
			if err := fn.CreateOrUpdateHandler(ctx, o, cl, scheme); err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, nil
	})
}

// copyTypedObject is used in order to provide an Operator for typed objects (GVK)
func copyTypedObject(object client.Object) client.Object {
	var obj client.Object
	obj = reflect.New(reflect.ValueOf(object).Elem().Type()).Interface().(client.Object)
	return obj
}

// copyUntypedObject is used in order to provide an Operator for untyped objects (GVK)
func copyUntypedObject(object client.Object, scheme *runtime.Scheme) client.Object {
	var obj client.Object
	obj = &unstructured.Unstructured{}
	gvk, err := apiutil.GVKForObject(object, scheme)
	if err != nil {
		return nil
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return obj
}
