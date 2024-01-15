package main

import (
	"context"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	lot_client "github.com/SchweizerischeBundesbahnen/lot/pkg/lot-client"
	"github.com/SchweizerischeBundesbahnen/lot/pkg/selector"

	"github.com/SchweizerischeBundesbahnen/lot/pkg/operator"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	log = logf.Log
)

func init() {
	opts := zap.Options{
		Development: true,
	}
	logf.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}
func main() {
	// TODO: Decide if we want to pass around a dedicated logger for the reconciler and its handlers?
	log := logf.Log.WithName("setup")
	log.Info("Start...")

	// Initialize operator with or without options
	// TODO: Documentation: Make clear that passing custom predicates allows the user to mute either update or delete
	// events, although they call onCreateOrUpdate.
	//o, err := operator.New(&v1.Secret{}, operator.WithPredicates(&customPredicates), operator.WithManagerOptions(&mgrOpts))
	//o, err := operator.New(&v1.Secret{}, operator.WithManagerOptions(&mgrOpts))
	//o, err := operator.New(&v1.Secret{})
	//o, err := operator.NewUntyped("", "v1", "Secret")
	o, err := operator.New(&v1.Secret{},
		operator.WithManagerOptions(&mgrOpts),
		operator.WithOwns(&v1.ServiceAccount{}, _defaultPredicate),
		operator.WithOwns(&v1.ConfigMap{}, _defaultPredicate))
	if err != nil {
		return
	}

	labelSelector := map[string]string{
		"foo":       "bar",
		"important": selector.KeyPresent(),
	}
	annotationSelector := map[string]string{
		"keepresource": selector.KeyAbsent(),
	}
	// Call event handlers, providing the respective reconcile handler func.
	// Handlers get only executed for resources that match the optional label and or annotation
	// selectors.
	o.OnCreateOrUpdate(createOrUpdateHandler, operator.WithLabels(labelSelector), operator.WithLabels(labelSelector))
	o.OnDelete(deleteHandler, operator.WithAnnotations(annotationSelector), operator.WithLabels(labelSelector))

	// Start the operator, this process builds the controller itself
	if err := o.Start(); err != nil {
		log.Error(err, "Problem starting operator")
		os.Exit(1)
	}
}

var mgrOpts = manager.Options{
	Cache:   cache.Options{DefaultNamespaces: map[string]cache.Config{"noexists": {}}},
	Metrics: server.Options{BindAddress: ":9090"},
}

var customPredicates = predicate.Funcs{
	UpdateFunc: func(event event.UpdateEvent) bool {
		log.Info("Custom: event for: ", "resource", event.ObjectOld.GetName())
		return true
	},
	DeleteFunc: func(event event.DeleteEvent) bool {
		log.Info("Custom: event for: ", "resource", event.Object.GetName())
		return true
	},
}

var _defaultPredicate = predicate.NewPredicateFuncs(func(object client.Object) bool {
	return true
})

// TODO: It is possible to return reconcile.Result (or a generic type of it..)
func createOrUpdateHandler(ctx context.Context, object client.Object, cl lot_client.Client, scheme *runtime.Scheme) error {
	log := logf.FromContext(ctx).WithName("onCreateOrUpdate")
	log.Info("reconciling created/updated object", "object", object.GetName(), "ns", object.GetNamespace())
	s := v1.Secret{}
	if object.GetName() == "delete-test-secret" {
		err := cl.Get(context.Background(), client.ObjectKey{Name: object.GetName(), Namespace: object.GetNamespace()}, &s)
		if err != nil {
			return err
		}
		log.Info("object found", s.GetName(), len(s.Data))
	}
	return nil
}
func deleteHandler(ctx context.Context, object client.Object, cl lot_client.Client, scheme *runtime.Scheme) error {
	log := logf.FromContext(ctx).WithName("onDelete")
	if object.GetDeletionTimestamp().IsZero() {
		return nil
	}
	log.Info("reconciling deleted object", "object", object.GetName(), "ns", object.GetNamespace())
	s := v1.Secret{}
	err := cl.Get(context.Background(), client.ObjectKey{Name: object.GetName(), Namespace: object.GetNamespace()}, &s)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("not found, so proceed doing something else...", "object", s.GetName())
		}
		return err
	}
	log.Info("curious.. the object is still here", "object", s.GetName(), "deletionTimestamp", s.GetDeletionTimestamp())
	return nil
}
