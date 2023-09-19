package operator

import (
	"github.com/SchweizerischeBundesbahnen/lot/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Operator knows how to call event handling operations for a specific event and Kubernetes resource
type Operator interface {
	OnCreateOrUpdate(handler reconcile.Handler, opts ...HandlerOption)
	OnDelete(handler reconcile.Handler, opts ...HandlerOption)
	Predicate() predicate.Predicate
	Build() error
	Start() error
}
