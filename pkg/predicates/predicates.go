package predicates

import (
	"github.com/SchweizerischeBundesbahnen/lot/pkg/selector"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// CreateOrUpdateByMetadata returns a predicate that filters Create and Update events based
// on labels or annotations.
func CreateOrUpdateByMetadata(labels map[string]string, annotations map[string]string) (predicate.Predicate, error) {
	s, err := selector.NewSelector(labels, annotations)
	if err != nil {
		return nil, err
	}
	result := predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			var labels, annotations map[string]string
			if labels = event.Object.GetLabels(); labels == nil {
				labels = map[string]string{}
			}
			if annotations = event.Object.GetAnnotations(); annotations == nil {
				annotations = map[string]string{}
			}
			return s.Matches(labels, annotations)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			var oldLabels, newLabels, oldAnnotations, newAnnotations map[string]string
			if oldLabels = event.ObjectOld.GetLabels(); oldLabels == nil {
				oldLabels = map[string]string{}
			}
			if oldAnnotations = event.ObjectOld.GetAnnotations(); oldAnnotations == nil {
				oldAnnotations = map[string]string{}
			}
			if newLabels = event.ObjectNew.GetLabels(); newLabels == nil {
				newLabels = map[string]string{}
			}
			if newAnnotations = event.ObjectNew.GetAnnotations(); newAnnotations == nil {
				newAnnotations = map[string]string{}
			}
			return s.Matches(oldLabels, oldAnnotations) || s.Matches(newLabels, newAnnotations)
		},
	}
	return result, nil
}

// DeleteByMetadata returns a predicate that filters Delete events based
// on labels or annotations.
func DeleteByMetadata(labels map[string]string, annotations map[string]string) (predicate.Predicate, error) {
	s, err := selector.NewSelector(labels, annotations)
	if err != nil {
		return nil, err
	}
	result := predicate.Funcs{
		DeleteFunc: func(event event.DeleteEvent) bool {
			var labels, annotations map[string]string
			if labels = event.Object.GetLabels(); labels == nil {
				labels = map[string]string{}
			}
			if annotations = event.Object.GetAnnotations(); annotations == nil {
				annotations = map[string]string{}
			}
			return s.Matches(labels, annotations)
		},
	}
	return result, nil
}

// Log returns a predicate that adds a logger to the given predicates so
// that processed events can be logged based on the loglevel. Events ignored
// by the input predicates are only logged when logIgnored is true. The return
// value of the input predicates is not changed.
func Log(log logr.Logger, logIgnored bool, p ...predicate.Predicate) predicate.Predicate {
	// predicate lists are implicitly added by the controller-runtime
	// (see https://github.com/kubernetes-sigs/controller-runtime/blob/v0.15.0/pkg/internal/source/event_handler.go#L79)
	// so we can do it here explicitly to add our log wrapper
	combined := predicate.And(p...)
	predicateLog := log.WithName("EventFilter")
	logWrapper := func(action string, decision bool, o client.Object) bool {
		eventLog := predicateLog.WithValues("event", action, "name", o.GetName(), "namespace", o.GetNamespace())

		if decision {
			eventLog.V(1).Info("Event accepted")
		}
		if !decision && logIgnored {
			eventLog.V(1).Info("Event ignored")
		}
		return decision
	}

	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return logWrapper("CREATE", combined.Create(e), e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return logWrapper("UPDATE", combined.Update(e), e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return logWrapper("DELETE", combined.Delete(e), e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return logWrapper("GENERIC", combined.Generic(e), e.Object)
		},
	}
}
