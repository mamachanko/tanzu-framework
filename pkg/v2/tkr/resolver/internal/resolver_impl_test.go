// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

func TestResolver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "tkr/resolver/internal Unit Tests")
}

const (
	k8s1_20_1 = "v1.20.1+vmware.1"
	k8s1_20_2 = "v1.20.2+vmware.1"
	k8s1_21_1 = "v1.21.1+vmware.1"
	k8s1_21_3 = "v1.21.3+vmware.1"
	k8s1_22_0 = "v1.22.0+vmware.1"
)

var k8sVersions = []string{k8s1_20_1, k8s1_20_2, k8s1_21_1, k8s1_21_3, k8s1_22_0}

const numOSImages = 50
const numTKRs = 10

var _ = Describe("Cache implementation", func() {
	var (
		osImages data.OSImages
		tkrs     data.TKRs

		osImagesByK8sVersion map[string]data.OSImages

		r *Resolver
	)

	BeforeEach(func() {
		osImages = testdata.GenOSImages(k8sVersions, numOSImages)
		osImagesByK8sVersion = testdata.SortOSImagesByK8sVersion(osImages)
		tkrs = testdata.GenTKRs(numTKRs, osImagesByK8sVersion)

		r = NewResolver()
	})

	var someOtherObject = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-config-map",
			Namespace: "some-namespace",
		},
		Data: map[string]string{},
	}

	BeforeEach(func() {
		for _, tkr := range tkrs {
			r.Add(tkr)
		}
		for _, osImage := range osImages {
			r.Add(osImage)
		}
		r.Add(someOtherObject)
	})

	Context("Add()", func() {
		It("should not matter if OSImages or TKRs are added first", func() {
			r1 := NewResolver()
			for _, osImage := range osImages {
				r1.Add(osImage)
			}
			for _, tkr := range tkrs {
				r1.Add(tkr)
			}
			r1.Add(someOtherObject)

			Expect(r1.cache.tkrs).To(Equal(r.cache.tkrs))
			Expect(r1.cache.osImages).To(Equal(r.cache.osImages))
			Expect(r1.cache.tkrToOSImages).To(Equal(r.cache.tkrToOSImages))
			Expect(r1.cache.osImageToTKRs).To(Equal(r.cache.osImageToTKRs))
		})

		It("should add TKRs and OSImages to the cache", func() {
			for tkrName, tkr := range tkrs {
				Expect(r.cache.tkrs).To(HaveKeyWithValue(tkrName, tkr))
				Expect(r.cache.tkrToOSImages).To(HaveKey(tkrName))
				shippedOSImages := r.cache.tkrToOSImages[tkrName]
				Expect(shippedOSImages).ToNot(BeNil())

				for _, osImageRef := range tkr.Spec.OSImages {
					osImageName := osImageRef.Name
					osImage := r.cache.osImages[osImageName]
					Expect(osImageName).ToNot(BeEmpty())
					Expect(osImage).ToNot(BeNil())

					Expect(shippedOSImages).To(HaveKeyWithValue(osImageName, osImage))
					Expect(r.cache.osImageToTKRs[osImageName]).To(HaveKeyWithValue(tkrName, tkr))
				}
			}
			for osImageName, osImage := range osImages {
				Expect(r.cache.osImages).To(HaveKeyWithValue(osImageName, osImage))
				Expect(r.cache.osImageToTKRs).To(HaveKey(osImageName))
				shippingTKRs := r.cache.osImageToTKRs[osImageName]
				Expect(shippingTKRs).ToNot(BeNil())

				for tkrName, tkr := range shippingTKRs {
					Expect(tkrName).ToNot(BeEmpty())
					Expect(tkr).ToNot(BeNil())
					Expect(tkr).To(Equal(r.cache.tkrs[tkrName]))
					Expect(r.cache.tkrToOSImages[tkrName]).To(HaveKeyWithValue(osImageName, osImage))
				}
			}
		})

		It("should not add other things to the cache", func() {
			Expect(r.cache.tkrs).ToNot(ContainElement(someOtherObject))
		})

		When("adding objects with DeletionTimestamp set", func() {
			var (
				tkrSubset     data.TKRs
				osImageSubset data.OSImages
			)

			BeforeEach(func() {
				tkrSubset = testdata.RandSubsetOfTKRs(tkrs)
				osImageSubset = testdata.RandSubsetOfOSImages(osImages)

				for _, tkr := range tkrSubset {
					Expect(r.cache.tkrs).To(HaveKeyWithValue(tkr.Name, tkr))
				}
				for _, osImage := range osImageSubset {
					Expect(r.cache.osImages).To(HaveKeyWithValue(osImage.Name, osImage))
				}
			})

			It("should remove them from the cache", func() {
				for _, tkr := range tkrSubset {
					tkr.DeletionTimestamp = &metav1.Time{Time: time.Now()}
					r.Add(tkr)
				}
				for _, osImage := range osImageSubset {
					osImage.DeletionTimestamp = &metav1.Time{Time: time.Now()}
					r.Add(osImage)
				}

				for _, tkr := range tkrSubset {
					Expect(r.cache.tkrs).ToNot(HaveKey(tkr.Name))
					Expect(r.cache.tkrToOSImages).ToNot(HaveKey(tkr.Name))
				}
				for _, osImage := range osImageSubset {
					Expect(r.cache.osImages).ToNot(HaveKey(osImage.Name))
					Expect(r.cache.osImageToTKRs).ToNot(HaveKey(osImage.Name))
				}
			})
		})
	})

	Context("Remove()", func() {
		var (
			tkrSubset     data.TKRs
			osImageSubset data.OSImages
		)

		BeforeEach(func() {
			tkrSubset = testdata.RandSubsetOfTKRs(tkrs)
			osImageSubset = testdata.RandSubsetOfOSImages(osImages)

			for _, tkr := range tkrSubset {
				Expect(r.cache.tkrs).To(HaveKeyWithValue(tkr.Name, tkr))
			}
			for _, osImage := range osImageSubset {
				Expect(r.cache.osImages).To(HaveKeyWithValue(osImage.Name, osImage))
			}
		})

		It("should remove them from the cache", func() {
			for _, tkr := range tkrSubset {
				r.Remove(tkr)
			}
			for _, osImage := range osImageSubset {
				r.Remove(osImage)
			}

			for _, tkr := range tkrSubset {
				Expect(r.cache.tkrs).ToNot(HaveKey(tkr.Name))
				Expect(r.cache.tkrToOSImages).ToNot(HaveKey(tkr.Name))
			}
			for _, osImage := range osImageSubset {
				Expect(r.cache.osImages).ToNot(HaveKey(osImage.Name))
				Expect(r.cache.osImageToTKRs).ToNot(HaveKey(osImage.Name))
			}
		})
	})

})

var _ = Describe("normalize(query)", func() {
	When("label selectors are empty", func() {
		var (
			initialQuery    = testdata.GenQueryAllForK8sVersion(k8sVersions[rand.Intn(len(k8sVersions))])
			normalizedQuery data.Query
		)

		BeforeEach(func() {
			normalizedQuery = normalize(initialQuery)
		})

		It("should add label requirements for the k8s version prefix", func() {
			assertOSImageQueryExpectations(normalizedQuery.ControlPlane, initialQuery.ControlPlane)
			for name, initialMDQuery := range initialQuery.MachineDeployments {
				Expect(normalizedQuery.MachineDeployments).To(HaveKey(name))
				assertOSImageQueryExpectations(normalizedQuery.MachineDeployments[name], initialMDQuery)
			}
		})
	})

})

func assertOSImageQueryExpectations(normalized, initial data.OSImageQuery) {
	Expect(normalized.K8sVersionPrefix).To(Equal(initial.K8sVersionPrefix))
	for _, selector := range []labels.Selector{normalized.TKRSelector, normalized.OSImageSelector} {
		Expect(selector.Matches(labels.Set{version.Label(initial.K8sVersionPrefix): ""})).To(BeTrue())
		Expect(selector.Matches(labels.Set{runv1.LabelIncompatible: ""})).To(BeFalse())
		Expect(selector.Matches(labels.Set{runv1.LabelDeactivated: ""})).To(BeFalse())
		Expect(selector.Matches(labels.Set{runv1.LabelInvalid: ""})).To(BeFalse())
	}
}

var _ = Describe("Resolve()", func() {
	var (
		osImages             data.OSImages
		osImagesByK8sVersion map[string]data.OSImages
		tkrs                 data.TKRs

		r *Resolver

		k8sVersion            string
		k8sVersionPrefix      string
		queryK8sVersionPrefix data.Query
	)

	BeforeEach(func() {
		osImages = testdata.GenOSImages(k8sVersions, numOSImages)
		osImagesByK8sVersion = testdata.SortOSImagesByK8sVersion(osImages)
		tkrs = testdata.GenTKRs(numTKRs, osImagesByK8sVersion)

		r = NewResolver()

		k8sVersion = testdata.ChooseK8sVersionFromTKRs(tkrs)
		k8sVersionPrefix = testdata.ChooseK8sVersionPrefix(k8sVersion)
		queryK8sVersionPrefix = testdata.GenQueryAllForK8sVersion(k8sVersionPrefix)
	})

	BeforeEach(func() {
		for _, tkr := range tkrs {
			r.Add(tkr)
		}
		for _, osImage := range osImages {
			r.Add(osImage)
		}
	})

	It("should resolve TKRs and OSImages for a version prefix", func() {
		result := r.Resolve(queryK8sVersionPrefix)

		assertOSImageResultExpectations(result.ControlPlane, queryK8sVersionPrefix.ControlPlane, k8sVersionPrefix)
		for name, osImageQuery := range queryK8sVersionPrefix.MachineDeployments {
			Expect(result.MachineDeployments).To(HaveKey(name))
			assertOSImageResultExpectations(result.MachineDeployments[name], osImageQuery, k8sVersionPrefix)
		}
	})
})

func assertOSImageResultExpectations(osImageResult data.OSImageResult, osImageQuery data.OSImageQuery, k8sVersionPrefix string) {
	Expect(version.Prefixes(osImageResult.K8sVersion)).To(HaveKey(k8sVersionPrefix))
	Expect(version.Prefixes(version.Label(osImageResult.TKRName))).To(HaveKey(version.Label(k8sVersionPrefix)))

	for k8sVersion, tkrs := range osImageResult.TKRsByK8sVersion {
		Expect(version.Prefixes(k8sVersion)).To(HaveKey(k8sVersionPrefix))
		Expect(tkrs).ToNot(BeEmpty())
		for tkrName, tkr := range tkrs {
			Expect(tkrName).To(Equal(tkr.Name))
			Expect(version.Prefixes(tkr.Spec.Version)).To(HaveKey(k8sVersionPrefix))
			Expect(version.Prefixes(tkr.Spec.Kubernetes.Version)).To(HaveKey(k8sVersionPrefix))
			Expect(osImageQuery.TKRSelector.Matches(labels.Set(tkr.Labels)))

			for osImageName, osImage := range osImageResult.OSImagesByTKR[tkrName] {
				Expect(osImageName).To(Equal(osImage.Name))
				Expect(version.Prefixes(osImage.Spec.KubernetesVersion)).To(HaveKey(k8sVersionPrefix))
				Expect(osImageQuery.OSImageSelector.Matches(labels.Set(osImage.Labels)))
			}
		}
	}
}
