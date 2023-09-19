package operator

import (
	"errors"
	"github.com/SchweizerischeBundesbahnen/lot/internal/defaults"
	lot_client "github.com/SchweizerischeBundesbahnen/lot/pkg/lot-client"
	"github.com/SchweizerischeBundesbahnen/lot/pkg/predicates"
	"github.com/SchweizerischeBundesbahnen/lot/pkg/reconcile"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ Operator = &operator{}

// operator is an operator.Operator that holds all the moving parts needed in order to build an controller-runtime controller
type operator struct {
	client     lot_client.Client
	object     client.Object
	controller controller.Controller
	// TODO: add a way to setup readyz and healthz endpoints for the operator
	manager           manager.Manager
	predicates        []predicate.Predicate
	customPredicates  []predicate.Predicate
	ownsInput         []OwnsInput
	reconcileHandlers *reconcile.HandlerFuncs
	errs              error
}

// New is a constructor function that creates a new instance of Operator.
// It takes an object of type client.Object and a variadic list of constructorOptions.
// It returns an Operator and an error, if any.
func New(object client.Object, opts ...ConstructorOption) (Operator, error) {
	var options constructorOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	var _ownsInput []OwnsInput
	if len(options.ownsInput) > 0 {
		_ownsInput = append(_ownsInput, options.ownsInput...)
	}

	mgr, err := defaults.InitManager(options.mgrOpts)
	if err != nil {
		return nil, err
	}

	var customPredicates []predicate.Predicate
	if options.predicates != nil {
		customPredicates = options.predicates
	}

	handlerFuncs := reconcile.HandlerFuncs{}

	cl := lot_client.New(mgr.GetClient())

	return &operator{
			client:            cl,
			object:            object,
			manager:           mgr,
			customPredicates:  customPredicates,
			ownsInput:         _ownsInput,
			reconcileHandlers: &handlerFuncs},
		nil
}

// NewUntyped is a constructor function that creates a new instance of Operator.
// It takes a string parameter for each GVK (group, version and kind) attributes of Kubernetes Object
// and a variadic list of constructorOptions.
// It returns an Operator and an error, if any.
func NewUntyped(group, version, kind string, opts ...ConstructorOption) (Operator, error) {
	var options constructorOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	var _ownsInput []OwnsInput
	if len(options.ownsInput) > 0 {
		_ownsInput = append(_ownsInput, options.ownsInput...)
	}

	mgr, err := defaults.InitManager(options.mgrOpts)
	if err != nil {
		return nil, err
	}

	var customPredicates []predicate.Predicate
	if options.predicates != nil {
		customPredicates = options.predicates
	}

	handlerFuncs := reconcile.HandlerFuncs{}

	cl := lot_client.New(mgr.GetClient())

	var object client.Object
	object = &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	object.GetObjectKind().SetGroupVersionKind(gvk)

	return &operator{
			client:            cl,
			object:            object,
			manager:           mgr,
			customPredicates:  customPredicates,
			ownsInput:         _ownsInput,
			reconcileHandlers: &handlerFuncs},
		nil
}

// OnCreateOrUpdate is a function that configures the predicate.CreateFunc and the reconcile.CreateOrUpdateHandler which are
// used to build the Operator's embedded controller.Controller. In this way the controller.Controller's Reconciler is able to
// handle create and update events accordingly
func (o *operator) OnCreateOrUpdate(fn reconcile.Handler, opts ...HandlerOption) {
	options := handlerOptions{
		labels:      map[string]string{},
		annotations: map[string]string{},
	}
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			o.errs = errors.Join(o.errs, err)
		}
	}
	// Filter out all non-delete events when using a delete handler.
	// See Predicate() for further details.
	defaultPredicate := predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return true },
		UpdateFunc:  func(event.UpdateEvent) bool { return true },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
	metadataPredicate, err := predicates.CreateOrUpdateByMetadata(options.labels, options.annotations)
	if err != nil {
		o.errs = errors.Join(o.errs, err)
	}
	handlerPredicate := predicate.And(defaultPredicate, metadataPredicate)
	o.predicates = append(o.predicates, handlerPredicate)

	o.reconcileHandlers.CreateOrUpdateHandler = fn
}

// OnDelete is a function that configures the predicate.DeleteFunc and the reconcileHandler.DeleteHandler which are
// used to build the Operator's embedded controller.Controller. In this way the controller.Controller's Reconciler is able to
// handle update events accordingly
func (o *operator) OnDelete(fn reconcile.Handler, opts ...HandlerOption) {
	options := handlerOptions{
		labels:      map[string]string{},
		annotations: map[string]string{},
	}
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			o.errs = errors.Join(o.errs, err)
		}
	}
	// Filter out all non-delete events when using a delete handler.
	// See Predicate() for further details.
	defaultPredicate := predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return false },
		UpdateFunc:  func(event.UpdateEvent) bool { return false },
		DeleteFunc:  func(event.DeleteEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
	metadataPredicate, err := predicates.DeleteByMetadata(options.labels, options.annotations)
	if err != nil {
		o.errs = errors.Join(o.errs, err)
	}
	handlerPredicate := predicate.And(defaultPredicate, metadataPredicate)
	o.predicates = append(o.predicates, handlerPredicate)

	o.reconcileHandlers.DeleteHandler = fn
}

// Start is a function that starts the embedded manager.Manager part of the Operator
func (o *operator) Start() error {
	if o.errs != nil {
		return o.errs
	}

	if err := o.Build(); err != nil {
		return err
	}

	if err := o.manager.Start(ctrl.SetupSignalHandler()); err != nil {
		return err
	}

	return nil
}

// Predicate constructs the predicate used to filter events for the primary resources
// i.e. the resources used in For().
func (o *operator) Predicate() predicate.Predicate {
	prcts := o.customPredicates

	// check if we have handler predicates and only integrate them if we have,
	// because a predicate.Or() with an empty list is always false which would effectively
	// block all events
	if len(o.predicates) > 0 {
		// The handler predicate ensures that only events to be handled by the OnDelete and OnCreateorUpdate
		// handlers are passed through. Each handler adds predicates that only permit events
		// for the specific handler and blocks all other events. By combining these predicates with a logical
		// or, we ensure that all events handled by any handler are processed.
		// Read this as: "The event either is intended for a handler, or it gets rejected"
		handlerPredicate := predicate.Or(o.predicates...)
		prcts = append(prcts, handlerPredicate)
	}

	// The operator predicate ensures that in addition to the handlerPredicate, all custom predicates
	// have to be fulfilled too.
	// Read this as: "The event must be intended for a handler and it must fulfill all custom predicates".
	operatorPredicate := predicate.And(prcts...)

	// wrap the operator predicate in a special logging predicate so we can enable event logging
	// by increasing the log level
	return predicates.Log(logf.Log, false, operatorPredicate)
}

// Build is a function that builds the embedded controller.Controller part of the Operator
// and registers the primary resource and its owned resources
// TODO: Add "Ows" and "Owns.Predicates" if possible
func (o *operator) Build() error {
	bldr := builder.
		ControllerManagedBy(o.manager).
		For(o.object, builder.WithPredicates(o.Predicate()))

	// TODO: catch nil predicate slice
	if len(o.ownsInput) > 0 {
		for _, input := range o.ownsInput {
			bldr.Owns(input.object, builder.WithPredicates(input.predicate))
		}
	}

	c, err := bldr.Build(o.reconcileFuncWithClient())
	if err != nil {
		return err
	}

	o.controller = c
	return nil
}

func (o *operator) reconcileFuncWithClient() reconcile.Reconciler {
	cl := o.client
	obj := o.object
	return reconcile.WithClient(cl, obj, o.manager.GetScheme(), o.reconcileHandlers)
}
