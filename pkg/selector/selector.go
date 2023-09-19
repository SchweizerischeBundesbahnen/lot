package selector

import (
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type state byte

const (
	keyAbsent state = iota
	keyPresent
)

var _ Selector = &selector{}

type Selector interface {
	Matches(labels map[string]string, annotations map[string]string) bool
	MatchesLabels(labels map[string]string) bool
	MatchesAnnotations(annotations map[string]string) bool
}

type selector struct {
	labels      map[string]string
	annotations map[string]string
}

// Control characters are never valid label or annotation values so we
// can use a string consisting of a 0-byte and a 1-byte to mark
// label / annotation keys that should exist/not exist, similar to the way
// python-kopf uses kopf.PRESENT and kopf.ABSENT (https://kopf.readthedocs.io/en/stable/filters/#metadata-filters).
// As there is no way in go to define an array/slice as const (and we need
// a slice to add a 0 byte to a string), we use functions to return a constant value.

// KeyAbsent returns a string that can be used as a label or annotation value
func KeyAbsent() string {
	return string([]state{keyAbsent})
}

// KeyPresent returns a string that can be used as a label or annotation value
func KeyPresent() string {
	return string([]state{keyPresent})
}

// NewSelector returns a new Selector that matches labels and annotations
func NewSelector(labels map[string]string, annotations map[string]string) (Selector, error) {
	var allErrs error
	// validate labels
	for k, v := range labels {
		// validate label keys
		path := field.ToPath()
		if err := validateKey(k, path.Child("key")); err != nil {
			allErrs = errors.Join(allErrs, err)
		}

		// validate label values, but skip our special "present" and "absent" values
		if (v == KeyPresent()) || (v == KeyAbsent()) {
			continue
		}
		if err := validateLabelValue(k, v, path.Child("values")); err != nil {
			allErrs = errors.Join(allErrs, err)
		}
	}
	// validate annotations
	// only validate keys here
	for k := range annotations {
		// validate annotation keys
		// there is no need to validate annotation values as annotation
		// values are not restricted
		path := field.ToPath()
		if err := validateKey(k, path.Child("key")); err != nil {
			allErrs = errors.Join(allErrs, err)
		}
	}
	return &selector{labels: labels, annotations: annotations}, allErrs
}

// Matches returns true if the given labels and annotations satisfy the requirements
// given when constructing the Selector.
func (s *selector) Matches(labels map[string]string, annotations map[string]string) bool {
	return s.MatchesLabels(labels) && s.MatchesAnnotations(annotations)
}

// MatchesLabels returns true if the given labels satisfy the label requirements
// given when constructing the Selector.
func (s *selector) MatchesLabels(labels map[string]string) bool {
	return matches(labels, s.labels)
}

// MatchesAnnotations returns true if the given labels satisfy the annotation requirements
// given when constructing the Selector.
func (s *selector) MatchesAnnotations(annotations map[string]string) bool {
	return matches(annotations, s.annotations)
}

func matches(data map[string]string, sel map[string]string) bool {
	if (sel == nil) || (data == nil) {
		return false
	}
	if len(sel) == 0 {
		return true
	}

	for k, v := range sel {
		switch v {
		case KeyAbsent():
			if _, ok := data[k]; ok {
				return false
			}
		case KeyPresent():
			if _, ok := data[k]; !ok {
				return false
			}
		default:
			if val, ok := data[k]; (!ok) || (v != val) {
				return false
			}
		}
	}
	return true
}

// source: https://github.com/kubernetes/apimachinery/blob/v0.26.1/pkg/labels/selector.go#L906
func validateKey(k string, path *field.Path) *field.Error {
	if errs := validation.IsQualifiedName(k); len(errs) != 0 {
		return field.Invalid(path, k, strings.Join(errs, "; "))
	}
	return nil
}

// source: https://github.com/kubernetes/apimachinery/blob/v0.26.1/pkg/labels/selector.go#L913
func validateLabelValue(k, v string, path *field.Path) *field.Error {
	if errs := validation.IsValidLabelValue(v); len(errs) != 0 {
		return field.Invalid(path.Key(k), v, strings.Join(errs, "; "))
	}
	return nil
}
