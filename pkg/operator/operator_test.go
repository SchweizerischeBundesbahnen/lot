package operator_test

import (
	"context"

	lotClient "github.com/SchweizerischeBundesbahnen/lot/pkg/lot-client"
	"github.com/SchweizerischeBundesbahnen/lot/pkg/operator"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// when spawning multiple operator instances, prevent
// them from creating listeners that could block each other
var disableHealthAndMetricEndpoint operator.ConstructorOption = operator.WithManagerOptions(&manager.Options{MetricsBindAddress: "0", HealthProbeBindAddress: "0"})

var _ = Describe("operator", func() {
	Describe("When creating a new operator", func() {

		It("should return success if given valid objects", func() {
			By("creating an operator without constructorOptions")
			o, err := operator.New(&v1.Secret{}, disableHealthAndMetricEndpoint)
			Expect(err).NotTo(HaveOccurred())
			Expect(o).NotTo(BeNil())
		})
		It("should return success if given valid objects", func() {
			By("creating an operator with constructorOptions")
			opts := manager.Options{MetricsBindAddress: ":9090"}
			o, err := operator.New(&v1.Secret{}, operator.WithManagerOptions(&opts))
			Expect(err).NotTo(HaveOccurred())
			Expect(o).NotTo(BeNil())
		})
		It("should return error if given twice the same opts function", func() {
			var o operator.Operator
			var err error

			{
				By("passing WithManagerOptions() twice")
				opts := manager.Options{MetricsBindAddress: ":9090"}
				o, err = operator.New(&v1.Secret{}, operator.WithManagerOptions(&opts), operator.WithManagerOptions(&opts))
				Expect(err).To(HaveOccurred())
				Expect(o).To(BeNil())
			}
		})
		It("should permit giving certain opts functions multiple times", func() {
			var o operator.Operator
			var err error

			{
				p := predicate.NewPredicateFuncs(func(object client.Object) bool {
					return true
				})
				By("passing WithCustomPredicate() twice")
				o, err = operator.New(&v1.Secret{}, disableHealthAndMetricEndpoint, operator.WithCustomPredicate(&p), operator.WithCustomPredicate(&p))
				Expect(err).ToNot(HaveOccurred())
				Expect(o).ToNot(BeNil())
			}
		})
		Describe("with handlers", func() {
			var o operator.Operator
			var err error
			BeforeEach(func() {
				o, err = operator.New(&v1.Secret{}, disableHealthAndMetricEndpoint)
				Expect(err).NotTo(HaveOccurred())
				Expect(o).NotTo(BeNil())
			})
			Describe("when defining an OnCreateOrUpdate handler", func() {
				It("should accept a handler function", func() {
					hdl := func(ctx context.Context, object client.Object, cl lotClient.Client, scheme *runtime.Scheme) error {
						return nil
					}
					o.OnCreateOrUpdate(hdl)
				})
				It("should accept a nil handler function", func() {
					o.OnCreateOrUpdate(nil)
				})
				It("should accept the WithAnnotations option", func() {
					o.OnCreateOrUpdate(nil, operator.WithAnnotations(map[string]string{"key": "value"}))
				})
				It("should accept the WithLabels option", func() {
					o.OnCreateOrUpdate(nil, operator.WithLabels(map[string]string{"key": "value"}))
				})
			})
			Describe("when defining a Delete handler", func() {
				It("should accept a handler function", func() {
					hdl := func(ctx context.Context, object client.Object, cl lotClient.Client, scheme *runtime.Scheme) error {
						return nil
					}
					o.OnDelete(hdl)
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
				})
				It("should accept a nil handler function", func() {
					o.OnDelete(nil)
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
				})
				It("should accept the WithAnnotations option", func() {
					o.OnDelete(nil, operator.WithAnnotations(map[string]string{"key": "value"}))
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
				})
				It("should accept the WithLabels option", func() {
					o.OnDelete(nil, operator.WithLabels(map[string]string{"key": "value"}))
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Describe("with predicates", func() {
			var o operator.Operator
			var prct predicate.Predicate
			var testPod, otherPod *v1.Pod
			var err error
			var testLabels, testAnnotations, otherLabels, otherAnnotations map[string]string
			var testCreateEvt, otherCreateEvt event.CreateEvent
			var testUpdateEvt, otherUpdateEvt event.UpdateEvent
			var testDeleteEvt, otherDeleteEvt event.DeleteEvent
			var testGenericEvt, otherGenericEvt event.GenericEvent
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
				testPod = &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   "biz",
						Name:        "baz",
						Labels:      testLabels,
						Annotations: testAnnotations,
					},
				}
				otherPod = &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   "biz",
						Name:        "baz",
						Labels:      otherLabels,
						Annotations: otherAnnotations,
					},
				}
				testCreateEvt = event.CreateEvent{Object: testPod}
				testUpdateEvt = event.UpdateEvent{ObjectOld: testPod, ObjectNew: testPod}
				testDeleteEvt = event.DeleteEvent{Object: testPod}
				testGenericEvt = event.GenericEvent{Object: testPod}

				otherCreateEvt = event.CreateEvent{Object: otherPod}
				otherUpdateEvt = event.UpdateEvent{ObjectOld: otherPod, ObjectNew: otherPod}
				otherDeleteEvt = event.DeleteEvent{Object: otherPod}
				otherGenericEvt = event.GenericEvent{Object: otherPod}

				o, err = operator.New(&v1.Secret{}, disableHealthAndMetricEndpoint)
				Expect(err).NotTo(HaveOccurred())
				Expect(o).NotTo(BeNil())
			})

			Describe("when only defining a CreateOrUpdate handler", func() {
				var opts []operator.HandlerOption
				JustBeforeEach(func() {
					o.OnCreateOrUpdate(nil, opts...)
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
					prct = o.Predicate()
				})
				Context("where no label/annotations selector is defined", func() {
					It("should only return true for all create/update events", func() {
						Expect(prct.Create(testCreateEvt)).To(BeTrue())
						Expect(prct.Update(testUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(testDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeTrue())
						Expect(prct.Update(otherUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
				Context("where a label selector is defined", func() {
					BeforeEach(func() {
						opts = []operator.HandlerOption{operator.WithLabels(testLabels)}
					})
					It("should only return true for all create/update events that have matching labels", func() {
						Expect(prct.Create(testCreateEvt)).To(BeTrue())
						Expect(prct.Update(testUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(testDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
				Context("where an annotation selector is defined", func() {
					BeforeEach(func() {
						opts = []operator.HandlerOption{operator.WithAnnotations(testAnnotations)}
					})
					It("should only return true for all create/update events that have matching annotations", func() {
						Expect(prct.Create(testCreateEvt)).To(BeTrue())
						Expect(prct.Update(testUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(testDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
				Context("where a label and an annotation selector is defined", func() {
					BeforeEach(func() {
						opts = []operator.HandlerOption{operator.WithLabels(testLabels), operator.WithAnnotations(testAnnotations)}
					})
					It("should only return true for all create/update events that have matching labels and annotations", func() {
						Expect(prct.Create(testCreateEvt)).To(BeTrue())
						Expect(prct.Update(testUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(testDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
			})

			Describe("when only defining a Delete handler", func() {
				var opts []operator.HandlerOption
				JustBeforeEach(func() {
					o.OnDelete(nil, opts...)
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
					prct = o.Predicate()
				})
				Context("where no label/annotations selector is defined", func() {
					It("should only return true for all delete events", func() {
						Expect(prct.Create(testCreateEvt)).To(BeFalse())
						Expect(prct.Update(testUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(testDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
				Context("where a label selector is defined", func() {
					BeforeEach(func() {
						opts = []operator.HandlerOption{operator.WithLabels(testLabels)}
					})
					It("should only return true for all delete events that have matching labels", func() {
						Expect(prct.Create(testCreateEvt)).To(BeFalse())
						Expect(prct.Update(testUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(testDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
				Context("where an annotation selector is defined", func() {
					BeforeEach(func() {
						opts = []operator.HandlerOption{operator.WithAnnotations(testAnnotations)}
					})
					It("should only return true for all delete events that have matching annotations", func() {
						Expect(prct.Create(testCreateEvt)).To(BeFalse())
						Expect(prct.Update(testUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(testDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
				Context("where a label and an annotation selector is defined", func() {
					BeforeEach(func() {
						opts = []operator.HandlerOption{operator.WithLabels(testLabels), operator.WithAnnotations(testAnnotations)}
					})
					It("should return true for all delete events that have matching labels and annotations", func() {
						Expect(prct.Create(testCreateEvt)).To(BeFalse())
						Expect(prct.Update(testUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(testDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
			})

			Describe("when defining a CreateOrUpdate and Delete handler", func() {
				var opts []operator.HandlerOption
				JustBeforeEach(func() {
					o.OnCreateOrUpdate(nil, opts...)
					o.OnDelete(nil, opts...)
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
					prct = o.Predicate()
				})
				Context("where no label/annotations selector is defined", func() {
					It("should only return true for all create/update/delete events", func() {
						Expect(prct.Create(testCreateEvt)).To(BeTrue())
						Expect(prct.Update(testUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(testDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeTrue())
						Expect(prct.Update(otherUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(otherDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
				Context("where a label and an annotation selector is defined", func() {
					BeforeEach(func() {
						opts = []operator.HandlerOption{operator.WithLabels(testLabels), operator.WithAnnotations(testAnnotations)}
					})
					It("should only return true for all create/update/delete events that have matching labels and annotations", func() {
						Expect(prct.Create(testCreateEvt)).To(BeTrue())
						Expect(prct.Update(testUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(testDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
			})

			Describe("when defining a custom predicate", func() {
				Context("where only a single custom predicate is defined", func() {
					It("should return the same result as the custom predicate", func() {
						prct := predicate.NewPredicateFuncs(func(object client.Object) bool {
							return true
						})
						o, err := operator.New(
							&v1.Secret{}, disableHealthAndMetricEndpoint, operator.WithCustomPredicate(prct))
						Expect(err).NotTo(HaveOccurred())
						Expect(o).NotTo(BeNil())
						Expect(o.Predicate().Create(testCreateEvt)).To(Equal(prct.Create(testCreateEvt)))
						Expect(o.Predicate().Update(testUpdateEvt)).To(Equal(prct.Update(testUpdateEvt)))
						Expect(o.Predicate().Delete(testDeleteEvt)).To(Equal(prct.Delete(testDeleteEvt)))
						Expect(o.Predicate().Generic(testGenericEvt)).To(Equal(prct.Generic(testGenericEvt)))

					})
				})
				Context("where multiple custom predicates are defined", func() {
					It("should combine the custom predicate results with 'and'", func() {
						inputs := [][]bool{
							{true, false},
							{true, true},
							{false, false},
						}
						for _, t := range inputs {
							p1 := predicate.NewPredicateFuncs(func(object client.Object) bool { return t[0] })
							p2 := predicate.NewPredicateFuncs(func(object client.Object) bool { return t[1] })
							pAnd := predicate.And(p1, p2)
							o, err := operator.New(&v1.Secret{}, disableHealthAndMetricEndpoint, operator.WithCustomPredicate(p1), operator.WithCustomPredicate(p2))
							Expect(err).NotTo(HaveOccurred())
							Expect(o).NotTo(BeNil())
							Expect(o.Predicate().Create(testCreateEvt)).To(Equal(pAnd.Create(testCreateEvt)))
							Expect(o.Predicate().Update(testUpdateEvt)).To(Equal(pAnd.Update(testUpdateEvt)))
							Expect(o.Predicate().Delete(testDeleteEvt)).To(Equal(pAnd.Delete(testDeleteEvt)))
							Expect(o.Predicate().Generic(testGenericEvt)).To(Equal(pAnd.Generic(testGenericEvt)))
						}
					})
				})
			})

			Describe("when combining handlers and custom predicates", func() {
				var customPrctResult bool
				JustBeforeEach(func() {
					customPrct := predicate.NewPredicateFuncs(func(object client.Object) bool {
						return customPrctResult
					})
					o, err = operator.New(&v1.Secret{}, operator.WithCustomPredicate(customPrct), disableHealthAndMetricEndpoint)
					Expect(err).NotTo(HaveOccurred())
					Expect(o).NotTo(BeNil())
					o.OnCreateOrUpdate(nil, operator.WithLabels(testLabels), operator.WithAnnotations(testAnnotations))
					o.OnDelete(nil, operator.WithLabels(testLabels), operator.WithAnnotations(testAnnotations))
					err := o.Build()
					Expect(err).ToNot(HaveOccurred())
					prct = o.Predicate()
				})

				Context("where the custom predicate returns true", func() {
					BeforeEach(func() {
						customPrctResult = true
					})
					It("should only return true for create/update/delete event that have matching labels/annotations", func() {
						Expect(prct.Create(testCreateEvt)).To(BeTrue())
						Expect(prct.Update(testUpdateEvt)).To(BeTrue())
						Expect(prct.Delete(testDeleteEvt)).To(BeTrue())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})

				Context("where the custom predicate returns false", func() {
					BeforeEach(func() {
						customPrctResult = false
					})
					It("should always return false", func() {
						Expect(prct.Create(testCreateEvt)).To(BeFalse())
						Expect(prct.Update(testUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(testDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(testGenericEvt)).To(BeFalse())

						Expect(prct.Create(otherCreateEvt)).To(BeFalse())
						Expect(prct.Update(otherUpdateEvt)).To(BeFalse())
						Expect(prct.Delete(otherDeleteEvt)).To(BeFalse())
						Expect(prct.Generic(otherGenericEvt)).To(BeFalse())
					})
				})
			})
		})
	})
})

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
})
