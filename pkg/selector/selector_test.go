package selector_test

import (
	"strings"

	"github.com/SchweizerischeBundesbahnen/lot/pkg/selector"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Selector", func() {
	Describe("When creating a new Selector", func() {
		var selLabels, selAnnotations map[string]string
		var err error
		var keySuccessCases, keyErrorCases, labelValueSuccessCases, labelValueErrorCases []string
		BeforeEach(func() {
			// values  taken from "func TestIsQualifiedName(t *testing.T)" in https://github.com/kubernetes/apimachinery/blob/v0.26.1/pkg/util/validation/validation_test.go#L247
			keySuccessCases = []string{
				"simple",
				"now-with-dashes",
				"1-starts-with-num",
				"1234",
				"simple/simple",
				"now-with-dashes/simple",
				"now-with-dashes/now-with-dashes",
				"now.with.dots/simple",
				"now-with.dashes-and.dots/simple",
				"1-num.2-num/3-num",
				"1234/5678",
				"1.2.3.4/5678",
				"Uppercase_Is_OK_123",
				"example.com/Uppercase_Is_OK_123",
				"requests.storage-foo",
				strings.Repeat("a", 63),
				strings.Repeat("a", 253) + "/" + strings.Repeat("b", 63),
			}
			// values taken from "func TestIsQualifiedName(t *testing.T)" in https://github.com/kubernetes/apimachinery/blob/v0.26.1/pkg/util/validation/validation_test.go#L247
			keyErrorCases = []string{
				"nospecialchars%^=@",
				"Tama-nui-te-rā.is.Māori.sun",
				"\\backslashes\\are\\bad",
				"-starts-with-dash",
				"ends-with-dash-",
				".starts.with.dot",
				"ends.with.dot.",
				strings.Repeat("a", 64), // over the limit
			}
			// values taken from "func TestIsValidLabelValue(t *testing.T)" in https://github.com/kubernetes/apimachinery/blob/v0.26.1/pkg/util/validation/validation_test.go#L292
			labelValueSuccessCases = []string{
				"simple",
				"now-with-dashes",
				"1-starts-with-num",
				"end-with-num-1",
				"1234",                  // only num
				strings.Repeat("a", 63), // to the limit
				"",                      // empty value
			}
			// values taken from "func TestIsValidLabelValue(t *testing.T)" in https://github.com/kubernetes/apimachinery/blob/v0.26.1/pkg/util/validation/validation_test.go#L292
			labelValueErrorCases = []string{
				"nospecialchars%^=@",
				"cantendwithadash-",
				"-cantstartwithadash-",
				"only/one/slash",
				"Example.com/abc",
				"example_com/abc",
				"example.com/",
				"/simple",
				strings.Repeat("a", 64),
				strings.Repeat("a", 254) + "/abc",
			}
		})
		Context("where valid label and annotation parameters were provided", func() {
			It("should succeed for labels with valid label keys", func() {
				for _, key := range keySuccessCases {
					selLabels = map[string]string{
						key: "value",
					}
					_, err = selector.NewSelector(selLabels, map[string]string{})
					Expect(err).ToNot(HaveOccurred())
				}
			})
			It("should succeed for labels with valid label values", func() {
				for _, value := range labelValueSuccessCases {
					selLabels = map[string]string{
						"key": value,
					}
					_, err = selector.NewSelector(selLabels, map[string]string{})
					Expect(err).ToNot(HaveOccurred())
				}
			})
			It("should succeed for annotations with valid annotation keys", func() {
				for _, key := range keySuccessCases {
					selAnnotations = map[string]string{
						key: "value",
					}
					_, err = selector.NewSelector(map[string]string{}, selAnnotations)
					Expect(err).ToNot(HaveOccurred())
				}
			})
			It("should succeed for annotations with valid label values", func() {
				for _, value := range labelValueSuccessCases {
					selAnnotations = map[string]string{
						"key": value,
					}
					_, err = selector.NewSelector(map[string]string{}, selAnnotations)
					Expect(err).ToNot(HaveOccurred())
				}
			})
			It("should succeed for annotations with invalid label values", func() {
				for _, value := range labelValueErrorCases {
					selAnnotations = map[string]string{
						"key": value,
					}
					_, err = selector.NewSelector(map[string]string{}, selAnnotations)
					Expect(err).ToNot(HaveOccurred())
				}
			})
		})
		Context("where invalid label and annotation parameters were provided", func() {
			It("should fail for labels with invalid label keys", func() {
				for _, key := range keyErrorCases {
					selLabels = map[string]string{
						key: "value",
					}
					_, err = selector.NewSelector(selLabels, map[string]string{})
					Expect(err).To(HaveOccurred())
				}
			})
			It("should fail for labels with invalid label values", func() {
				for _, value := range labelValueErrorCases {
					selLabels = map[string]string{
						"key": value,
					}
					_, err = selector.NewSelector(selLabels, map[string]string{})
					Expect(err).To(HaveOccurred())
				}
			})
			It("should fail for annotations with invalid annotations keys", func() {
				for _, key := range keyErrorCases {
					selAnnotations = map[string]string{
						key: "value",
					}
					_, err = selector.NewSelector(map[string]string{}, selAnnotations)
					Expect(err).To(HaveOccurred())
				}
			})
		})
	})
	Describe("When using a Selector", func() {
		var selLabels, selAnnotations, otherLabels, otherAnnotations map[string]string
		var sel selector.Selector
		var err error
		JustBeforeEach(func() {
			otherLabels = map[string]string{"otherlabelkey": "otherlabelvalue"}
			otherAnnotations = map[string]string{"otherannotationkey": "otherannotationvalue"}
			sel, err = selector.NewSelector(selLabels, selAnnotations)
			Expect(err).ToNot(HaveOccurred())
			Expect(sel).NotTo(BeNil())
		})
		Context("where non-empty label and annotation parameters were provided", func() {
			BeforeEach(func() {
				selLabels = map[string]string{"labelkey": "labelvalue"}
				selAnnotations = map[string]string{"annotationkey": "annotationvalue"}
			})
			Describe("when only matching labels", func() {
				It("should match the labels given as input", func() {
					Expect(sel.MatchesLabels(selLabels)).To(BeTrue())
				})
				It("should not match other labels", func() {
					Expect(sel.MatchesLabels(otherLabels)).To(BeFalse())
				})
			})
			Describe("when only matching annotations", func() {
				It("should match the annotations given as input", func() {
					Expect(sel.MatchesAnnotations(selAnnotations)).To(BeTrue())
				})
				It("should not match other annotations", func() {
					Expect(sel.MatchesAnnotations(otherAnnotations)).To(BeFalse())
				})
			})
			Describe("when matching labels and annotations", func() {
				It("should match the labels and annotations given as input", func() {
					Expect(sel.Matches(selLabels, selAnnotations)).To(BeTrue())
				})
				It("should not match other labels and annotations", func() {
					Expect(sel.Matches(otherLabels, otherAnnotations)).To(BeFalse())
				})
				It("should not match other annotations", func() {
					Expect(sel.Matches(selLabels, otherAnnotations)).To(BeFalse())
				})
				It("should not match other labels", func() {
					Expect(sel.Matches(otherLabels, selAnnotations)).To(BeFalse())
				})
			})
		})
		Context("where nil label and annotation parameters were provided", func() {
			BeforeEach(func() {
				selLabels = nil
				selAnnotations = nil
			})
			Describe("when only matching labels", func() {
				It("should not match the labels given as input", func() {
					Expect(sel.MatchesLabels(selLabels)).To(BeFalse())
				})
				It("should not match other labels", func() {
					Expect(sel.MatchesLabels(otherLabels)).To(BeFalse())
				})
			})
			Describe("when only matching annotations", func() {
				It("should not match the annotations given as input", func() {
					Expect(sel.MatchesAnnotations(selAnnotations)).To(BeFalse())
				})
				It("should not match other annotations", func() {
					Expect(sel.MatchesAnnotations(otherAnnotations)).To(BeFalse())
				})
			})
			Describe("when matching labels and annotations", func() {
				It("should not match the labels and annotations given as input", func() {
					Expect(sel.Matches(selLabels, selAnnotations)).To(BeFalse())
				})
				It("should not match other labels and annotations", func() {
					Expect(sel.Matches(otherLabels, otherAnnotations)).To(BeFalse())
				})
				It("should not match other annotations", func() {
					Expect(sel.Matches(selLabels, otherAnnotations)).To(BeFalse())
				})
				It("should not match other labels", func() {
					Expect(sel.Matches(otherLabels, selAnnotations)).To(BeFalse())
				})
			})
		})
		Context("where empty label and annotation parameters were provided", func() {
			BeforeEach(func() {
				selLabels = map[string]string{}
				selAnnotations = map[string]string{}
			})
			Describe("when only matching labels", func() {
				It("should match the labels given as input", func() {
					Expect(sel.MatchesLabels(selLabels)).To(BeTrue())
				})
				It("should match other labels", func() {
					Expect(sel.MatchesLabels(otherLabels)).To(BeTrue())
				})
				It("should not match nil", func() {
					Expect(sel.MatchesLabels(nil)).To(BeFalse())
				})
			})
			Describe("when only matching annotations", func() {
				It("should match the annotations given as input", func() {
					Expect(sel.MatchesAnnotations(selAnnotations)).To(BeTrue())
				})
				It("should match other annotations", func() {
					Expect(sel.MatchesAnnotations(otherAnnotations)).To(BeTrue())
				})
				It("should not match nil", func() {
					Expect(sel.MatchesAnnotations(nil)).To(BeFalse())
				})
			})
			Describe("when matching labels and annotations", func() {
				It("should match the labels and annotations given as input", func() {
					Expect(sel.Matches(selLabels, selAnnotations)).To(BeTrue())
				})
				It("should match other labels and annotations", func() {
					Expect(sel.Matches(otherLabels, otherAnnotations)).To(BeTrue())
				})
				It("should match other annotations", func() {
					Expect(sel.Matches(selLabels, otherAnnotations)).To(BeTrue())
				})
				It("should match other labels", func() {
					Expect(sel.Matches(otherLabels, selAnnotations)).To(BeTrue())
				})
				It("should not match nil labels and annotations", func() {
					Expect(sel.Matches(nil, nil)).To(BeFalse())
				})
				It("should not match nil labels", func() {
					Expect(sel.Matches(nil, selAnnotations)).To(BeFalse())
				})
				It("should not match nil labels", func() {
					Expect(sel.Matches(selLabels, nil)).To(BeFalse())
				})
			})
		})
		Describe("when key presence / absence should be matched", func() {
			var withKey, withoutKey map[string]string
			BeforeEach(func() {
				withKey = map[string]string{"exists": "ispresent"}
				withoutKey = map[string]string{"otherkey": "something"}
			})

			Context("where label and annotation parameters were provided that should have a certain key", func() {
				BeforeEach(func() {
					selLabels = map[string]string{
						"exists": selector.KeyPresent(),
					}
					selAnnotations = map[string]string{
						"exists": selector.KeyPresent(),
					}
				})
				Describe("when only matching labels", func() {
					It("should match labels that have the key", func() {
						Expect(sel.MatchesLabels(withKey)).To(BeTrue())
					})
					It("should not match labels that do not have the key", func() {
						Expect(sel.MatchesLabels(withoutKey)).To(BeFalse())
					})
				})
				Describe("when only matching annotations", func() {
					It("should match annotations that have the key", func() {
						Expect(sel.MatchesAnnotations(withKey)).To(BeTrue())
					})
					It("should not match annotations that do not have the key", func() {
						Expect(sel.MatchesAnnotations(withoutKey)).To(BeFalse())
					})
				})
				Describe("when matching labels and annotations", func() {
					It("should match labels and annotations that have the key", func() {
						Expect(sel.Matches(withKey, withKey)).To(BeTrue())
					})
					It("should not match labels and annotations that do not have the key", func() {
						Expect(sel.Matches(withoutKey, withoutKey)).To(BeFalse())
					})
					It("should not match annotations that do not have the key", func() {
						Expect(sel.Matches(withKey, withoutKey)).To(BeFalse())
					})
					It("should not match labels that do not have the key", func() {
						Expect(sel.Matches(withoutKey, withKey)).To(BeFalse())
					})
				})
			})
			Context("where label and annotation parameters were provided that should not have a certain key", func() {
				BeforeEach(func() {
					selLabels = map[string]string{
						"exists": selector.KeyAbsent(),
					}
					selAnnotations = map[string]string{
						"exists": selector.KeyAbsent(),
					}
				})
				Describe("when only matching labels", func() {
					It("should match labels that do not have the key", func() {
						Expect(sel.MatchesLabels(withoutKey)).To(BeTrue())
					})
					It("should not match labels that have the key", func() {
						Expect(sel.MatchesLabels(withKey)).To(BeFalse())
					})
				})
				Describe("when only matching annotations", func() {
					It("should match annotations that do not have the key", func() {
						Expect(sel.MatchesAnnotations(withoutKey)).To(BeTrue())
					})
					It("should not match annotations that have the key", func() {
						Expect(sel.MatchesAnnotations(withKey)).To(BeFalse())
					})
				})
				Describe("when matching labels and annotations", func() {
					It("should match labels and annotations that do not have the key", func() {
						Expect(sel.Matches(withoutKey, withoutKey)).To(BeTrue())
					})
					It("should not match labels and annotations that have the key", func() {
						Expect(sel.Matches(withKey, withKey)).To(BeFalse())
					})
					It("should not match annotations that have the key", func() {
						Expect(sel.Matches(withoutKey, withKey)).To(BeFalse())
					})
					It("should not match labels that have the key", func() {
						Expect(sel.Matches(withKey, withoutKey)).To(BeFalse())
					})
				})
			})
		})
	})
})
