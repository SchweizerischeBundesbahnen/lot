package predicates_test

import (
	"github.com/SchweizerischeBundesbahnen/lot/pkg/predicates"
	"github.com/SchweizerischeBundesbahnen/lot/pkg/selector"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ = Describe("Predicates", func() {
	Describe("When checking a byMetadata predicate", func() {
		var testLabels, testAnnotations, otherLabels, otherAnnotations map[string]string
		BeforeEach(func() {
			testLabels = map[string]string{
				"labelkey": "labelvalue",
			}
			testAnnotations = map[string]string{
				"annotationkey": "annotationvalue",
			}
			otherLabels = map[string]string{
				"otherlabelkey": "otherlabelvalue",
			}
			otherAnnotations = map[string]string{
				"otherannotationkey": "otherannotationvalue",
			}
		})

		Describe("when checking a CreateOrUpdateByMetadata predicate", func() {
			var labels, annotations map[string]string
			var instance predicate.Predicate
			var err error
			JustBeforeEach(func() {
				instance, err = predicates.CreateOrUpdateByMetadata(labels, annotations)
				Expect(instance).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
			Context("where only labels should be considered", func() {
				BeforeEach(func() {
					labels = testLabels
					annotations = map[string]string{}
				})
				It("should return true for create/update events for objects with the expected labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
							Labels:    testLabels,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeTrue())
					Expect(instance.Update(updateEvt)).To(BeTrue())
				})
				It("should return false for create/update events for objects missing the expected labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
							Labels:    otherLabels,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return true for update events where either the old or new object has the expected labels", func() {
					podOk := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
							Labels:    testLabels,
						},
					}
					podNok := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
							Labels:    otherLabels,
						},
					}
					onlyOldEvt := event.UpdateEvent{ObjectOld: podOk, ObjectNew: podNok}
					onlyNewEvt := event.UpdateEvent{ObjectOld: podNok, ObjectNew: podOk}
					Expect(instance.Update(onlyOldEvt)).To(BeTrue())
					Expect(instance.Update(onlyNewEvt)).To(BeTrue())
				})
				It("should return false for create/update events for objects without labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return true for any annotation values", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Labels:      labels,
								Annotations: a,
							},
						}
						createEvt := event.CreateEvent{Object: pod}
						updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
						Expect(instance.Create(createEvt)).To(BeTrue())
						Expect(instance.Update(updateEvt)).To(BeTrue())
					}
				})
				It("should return true for all delete/generic events", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "biz",
								Name:      "baz",
								Labels:    a,
							},
						}
						deleteEvt := event.DeleteEvent{Object: pod}
						genericEvt := event.GenericEvent{Object: pod}
						Expect(instance.Delete(deleteEvt)).To(BeTrue())
						Expect(instance.Generic(genericEvt)).To(BeTrue())
					}
				})
			})
			Context("where only annotations should be considered", func() {
				BeforeEach(func() {
					labels = map[string]string{}
					annotations = testAnnotations
				})
				It("should return true for create/update events for objects with the expected annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeTrue())
					Expect(instance.Update(updateEvt)).To(BeTrue())
				})
				It("should return false for create/update events for objects missing the expected annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return true for update events where either the old or new object has the expected annotations", func() {
					podOk := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
						},
					}
					podNok := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
						},
					}
					onlyOldEvt := event.UpdateEvent{ObjectOld: podOk, ObjectNew: podNok}
					onlyNewEvt := event.UpdateEvent{ObjectOld: podNok, ObjectNew: podOk}
					Expect(instance.Update(onlyOldEvt)).To(BeTrue())
					Expect(instance.Update(onlyNewEvt)).To(BeTrue())
				})
				It("should return false for create/update events for objects without annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return true for any label values", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Labels:      a,
								Annotations: annotations,
							},
						}
						createEvt := event.CreateEvent{Object: pod}
						updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
						Expect(instance.Create(createEvt)).To(BeTrue())
						Expect(instance.Update(updateEvt)).To(BeTrue())
					}
				})
				It("should return true for all delete/generic events", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Annotations: a,
							},
						}
						deleteEvt := event.DeleteEvent{Object: pod}
						genericEvt := event.GenericEvent{Object: pod}
						Expect(instance.Delete(deleteEvt)).To(BeTrue())
						Expect(instance.Generic(genericEvt)).To(BeTrue())
					}
				})
			})

			Context("where labels and annotations should be considered", func() {
				BeforeEach(func() {
					labels = testLabels
					annotations = testAnnotations
				})
				It("should return true for create/update events for objects with the expected labels/annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
							Labels:      testLabels,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeTrue())
					Expect(instance.Update(updateEvt)).To(BeTrue())
				})
				It("should return false for create/update events for objects missing the expected labels/annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
							Labels:      otherLabels,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return false for create/update events for objects missing the expected labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
							Labels:      otherLabels,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return false for create/update events for objects missing the expected annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
							Labels:      testLabels,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return true for update events where either the old or new object has the expected labels/annotations", func() {
					podOk := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
							Labels:      testLabels,
						},
					}
					podNok := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
							Labels:      otherLabels,
						},
					}
					onlyOldEvt := event.UpdateEvent{ObjectOld: podOk, ObjectNew: podNok}
					onlyNewEvt := event.UpdateEvent{ObjectOld: podNok, ObjectNew: podOk}
					Expect(instance.Update(onlyOldEvt)).To(BeTrue())
					Expect(instance.Update(onlyNewEvt)).To(BeTrue())
				})
				It("should return false for create/update events for objects without labels/annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
				It("should return true for all delete/generic events", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Annotations: a,
								Labels:      a,
							},
						}
						deleteEvt := event.DeleteEvent{Object: pod}
						genericEvt := event.GenericEvent{Object: pod}
						Expect(instance.Delete(deleteEvt)).To(BeTrue())
						Expect(instance.Generic(genericEvt)).To(BeTrue())
					}
				})
			})
			Context("where special presence selector values are used", func() {
				BeforeEach(func() {
					labels = map[string]string{"labelkey": selector.KeyPresent()}
					annotations = map[string]string{"annotationkey": selector.KeyPresent()}
				})
				It("should return true for labels/annotations where the keys are present", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      testLabels,
							Annotations: testAnnotations,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeTrue())
					Expect(instance.Update(updateEvt)).To(BeTrue())
				})
				It("should return false for labels/annotations where the keys are absent", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      otherLabels,
							Annotations: otherAnnotations,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
			})
			Context("where special absence selector values are used", func() {
				BeforeEach(func() {
					labels = map[string]string{"labelkey": selector.KeyAbsent()}
					annotations = map[string]string{"annotationkey": selector.KeyAbsent()}
				})
				It("should return true for labels/annotations where the keys are absent", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      otherLabels,
							Annotations: otherAnnotations,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeTrue())
					Expect(instance.Update(updateEvt)).To(BeTrue())
				})
				It("should return false for labels/annotations where the keys are present", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      testLabels,
							Annotations: testAnnotations,
						},
					}
					createEvt := event.CreateEvent{Object: pod}
					updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
					Expect(instance.Create(createEvt)).To(BeFalse())
					Expect(instance.Update(updateEvt)).To(BeFalse())
				})
			})
		})
		Describe("when checking a DeleteByMetadata predicate", func() {
			var labels, annotations map[string]string
			var instance predicate.Predicate
			var err error
			JustBeforeEach(func() {
				instance, err = predicates.DeleteByMetadata(labels, annotations)
				Expect(instance).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
			Context("where only labels should be considered", func() {
				BeforeEach(func() {
					labels = testLabels
					annotations = map[string]string{}
				})
				It("should return true for delete events for objects with the expected labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
							Labels:    testLabels,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeTrue())
				})
				It("should return false for delete events for objects missing the expected labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
							Labels:    otherLabels,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return false for delete events for objects without labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return true for any annotation values", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Labels:      labels,
								Annotations: a,
							},
						}
						evt := event.DeleteEvent{Object: pod}
						Expect(instance.Delete(evt)).To(BeTrue())
					}
				})
				It("should return true for all create/update/generic events", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "biz",
								Name:      "baz",
								Labels:    a,
							},
						}
						createEvt := event.CreateEvent{Object: pod}
						updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
						genericEvt := event.GenericEvent{Object: pod}
						Expect(instance.Create(createEvt)).To(BeTrue())
						Expect(instance.Update(updateEvt)).To(BeTrue())
						Expect(instance.Generic(genericEvt)).To(BeTrue())
					}
				})
			})
			Context("where only annotations should be considered", func() {
				BeforeEach(func() {
					labels = map[string]string{}
					annotations = testAnnotations
				})
				It("should return true for delete events for objects with the expected annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeTrue())
				})
				It("should return false for delete events for objects missing the expected annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return false for delete events for objects without annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return true for any label values", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Labels:      a,
								Annotations: annotations,
							},
						}
						evt := event.DeleteEvent{Object: pod}
						Expect(instance.Delete(evt)).To(BeTrue())
					}
				})
				It("should return true for all create/update/generic events", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Annotations: a,
							},
						}
						createEvt := event.CreateEvent{Object: pod}
						updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
						genericEvt := event.GenericEvent{Object: pod}
						Expect(instance.Create(createEvt)).To(BeTrue())
						Expect(instance.Update(updateEvt)).To(BeTrue())
						Expect(instance.Generic(genericEvt)).To(BeTrue())
					}
				})
			})

			Context("where labels and annotations should be considered", func() {
				BeforeEach(func() {
					labels = testLabels
					annotations = testAnnotations
				})
				It("should return true for delete events for objects with the expected labels/annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
							Labels:      testLabels,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeTrue())
				})
				It("should return false for delete events for objects missing the expected labels/annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
							Labels:      otherLabels,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return false for delete events for objects missing the expected labels", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: testAnnotations,
							Labels:      otherLabels,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return false for delete events for objects missing the expected annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Annotations: otherAnnotations,
							Labels:      testLabels,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return false for delete events for objects without labels/annotations", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "biz",
							Name:      "baz",
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
				It("should return true for all create/update/generic events", func() {
					accepted := []map[string]string{
						{},
						{"key": "value"},
						nil,
					}
					for _, a := range accepted {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace:   "biz",
								Name:        "baz",
								Annotations: a,
								Labels:      a,
							},
						}
						createEvt := event.CreateEvent{Object: pod}
						updateEvt := event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
						genericEvt := event.GenericEvent{Object: pod}
						Expect(instance.Create(createEvt)).To(BeTrue())
						Expect(instance.Update(updateEvt)).To(BeTrue())
						Expect(instance.Generic(genericEvt)).To(BeTrue())
					}
				})
			})
			Context("where special presence selector values are used", func() {
				BeforeEach(func() {
					labels = map[string]string{"labelkey": selector.KeyPresent()}
					annotations = map[string]string{"annotationkey": selector.KeyPresent()}
				})
				It("should return true for labels/annotations where the keys are present", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      testLabels,
							Annotations: testAnnotations,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeTrue())
				})
				It("should return false for labels/annotations where the keys are absent", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      otherLabels,
							Annotations: otherAnnotations,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
			})
			Context("where special absence selector values are used", func() {
				BeforeEach(func() {
					labels = map[string]string{"labelkey": selector.KeyAbsent()}
					annotations = map[string]string{"annotationkey": selector.KeyAbsent()}
				})
				It("should return true for labels/annotations where the keys are absent", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      otherLabels,
							Annotations: otherAnnotations,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeTrue())
				})
				It("should return false for labels/annotations where the keys are present", func() {
					pod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Namespace:   "biz",
							Name:        "baz",
							Labels:      testLabels,
							Annotations: testAnnotations,
						},
					}
					evt := event.DeleteEvent{Object: pod}
					Expect(instance.Delete(evt)).To(BeFalse())
				})
			})
		})
	})
	Describe("When checking a Log predicate", func() {
		var pod *corev1.Pod
		var createEvt event.CreateEvent
		var updateEvt event.UpdateEvent
		var deleteEvt event.DeleteEvent
		var genericEvt event.GenericEvent
		BeforeEach(func() {
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: "biz", Name: "baz"},
			}
			createEvt = event.CreateEvent{Object: pod}
			updateEvt = event.UpdateEvent{ObjectOld: pod, ObjectNew: pod}
			deleteEvt = event.DeleteEvent{Object: pod}
			genericEvt = event.GenericEvent{Object: pod}
		})

		// as the Log predicate is just a wrapper, it should not influence
		// the result of any predicate it wraps
		It("should pass through the result of a single predicate", func() {
			for _, t := range []bool{true, false} {
				input := predicate.NewPredicateFuncs(func(object client.Object) bool { return t })
				instance := predicates.Log(logf.Log, false, input)
				Expect(instance.Create(createEvt)).To(Equal(t))
				Expect(instance.Update(updateEvt)).To(Equal(t))
				Expect(instance.Delete(deleteEvt)).To(Equal(t))
				Expect(instance.Generic(genericEvt)).To(Equal(t))
			}
		})
		// ensure the predicate handles multiple input predicates the same way
		// as controller runtime handles multiple predicates
		It("should combine multiple predicates with 'and'", func() {
			inputs := [][]bool{
				{true, false},
				{true, true},
				{false, false},
			}
			for _, t := range inputs {
				p1 := predicate.NewPredicateFuncs(func(object client.Object) bool { return t[0] })
				p2 := predicate.NewPredicateFuncs(func(object client.Object) bool { return t[1] })
				instance := predicates.Log(logf.Log, false, p1, p2)
				Expect(instance.Create(createEvt)).To(Equal(t[0] && t[1]))
				Expect(instance.Update(updateEvt)).To(Equal(t[0] && t[1]))
				Expect(instance.Delete(deleteEvt)).To(Equal(t[0] && t[1]))
				Expect(instance.Generic(genericEvt)).To(Equal(t[0] && t[1]))
			}
		})
	})
})
