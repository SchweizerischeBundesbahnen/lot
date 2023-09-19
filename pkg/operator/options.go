package operator

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type constructorOptions struct {
	mgrOpts    *manager.Options
	predicates []predicate.Predicate
	ownsInput  []OwnsInput
}

type OwnsInput struct {
	object    client.Object
	predicate predicate.Predicate
}

type ConstructorOption func(options *constructorOptions) error

func WithCustomPredicate(prct predicate.Predicate) ConstructorOption {
	return func(opts *constructorOptions) error {
		opts.predicates = append(opts.predicates, prct)
		return nil
	}
}

func WithManagerOptions(mgrOpts *manager.Options) ConstructorOption {
	return func(opts *constructorOptions) error {
		if opts.mgrOpts != nil {
			return fmt.Errorf("WithManager(...) should only be called once")
		}
		if mgrOpts != nil {
			opts.mgrOpts = mgrOpts
		}
		return nil
	}
}

func WithOwns(object client.Object, filter predicate.Predicate) ConstructorOption {
	return func(opts *constructorOptions) error {
		input := OwnsInput{object: object, predicate: filter}
		opts.ownsInput = append(opts.ownsInput, input)
		return nil
	}
}

type handlerOptions struct {
	labels      map[string]string
	annotations map[string]string
}

type HandlerOption func(options *handlerOptions) error

// WithAnnotations sets the annotations used to filter events
// for the handler. Calling it multiple times merges the values.
func WithAnnotations(annotations map[string]string) HandlerOption {
	return func(opts *handlerOptions) error {
		if opts.annotations == nil {
			opts.annotations = make(map[string]string)
		}
		for k, v := range annotations {
			opts.annotations[k] = v
		}
		return nil
	}
}

// WithLabels sets the labels used to filter events
// for the handler. Calling it multiple times merges the values.
func WithLabels(labels map[string]string) HandlerOption {
	return func(opts *handlerOptions) error {
		if opts.labels == nil {
			opts.labels = make(map[string]string)
		}
		for k, v := range labels {
			opts.labels[k] = v
		}
		return nil
	}
}
