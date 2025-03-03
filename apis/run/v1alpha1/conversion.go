// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	utilconversion "github.com/vmware-tanzu/tanzu-framework/util/conversion"
)

const (
	etcdContainerImageName    = "etcd"
	corednsContainerImageName = "coredns"
	pauseContainerImageName   = "pause"
)

func (spoke *TanzuKubernetesRelease) ConvertTo(hubRaw conversion.Hub) error {
	hub := hubRaw.(*v1alpha3.TanzuKubernetesRelease)

	if err := autoConvert_v1alpha1_TanzuKubernetesRelease_To_v1alpha3_TanzuKubernetesRelease(spoke, hub, nil); err != nil {
		return nil
	}

	// Some Hub types not present in spoke, and might be stored in spoke annotations.
	// Restore data stored in spoke's annotations.
	restoredHub := &v1alpha3.TanzuKubernetesRelease{}
	if ok, err := utilconversion.UnmarshalData(spoke, restoredHub); err != nil {
		return err
	} else if ok {
		// OK means there is data in annotations that need to be restored.
		hub.Spec.BootstrapPackages = restoredHub.Spec.BootstrapPackages
		hub.Spec.OSImages = restoredHub.Spec.OSImages
	}

	// Store the entire spoke in Hub's annotations to prevent loss of incompatible spoke types.
	if err := utilconversion.MarshalData(spoke, hub); err != nil {
		return err
	}

	return nil
}

//nolint:stylecheck // Much easier to read when the receiver is named as spoke/src/dest.
func (spoke *TanzuKubernetesRelease) ConvertFrom(hubRaw conversion.Hub) error {
	hub := hubRaw.(*v1alpha3.TanzuKubernetesRelease)

	if err := autoConvert_v1alpha3_TanzuKubernetesRelease_To_v1alpha1_TanzuKubernetesRelease(hub, spoke, nil); err != nil {
		return err
	}

	// From hub, get missing spoke fields from annotations.
	restored := &TanzuKubernetesRelease{}
	if ok, err := utilconversion.UnmarshalData(hub, restored); err != nil {
		return err
	} else if ok {
		// Annotations contain data, restore the relevant ones.
		spoke.Spec.NodeImageRef = restored.Spec.NodeImageRef
		if restored.Spec.Images != nil {
			spoke.Spec.Images = restored.Spec.Images
		}
	}

	// Store hub's missing fields in spoke's annotations.
	if err := utilconversion.MarshalData(hub, spoke); err != nil {
		return err
	}

	return nil
}

// Convert_v1alpha1_TanzuKubernetesReleaseSpec_To_v1alpha3_TanzuKubernetesReleaseSpec  will convert the compatible types in TanzuKubernetesReleaseSpec v1alpha1 to v1alpha3 equivalent.
// nolint:revive,stylecheck // Generated conversion stubs have underscores in function names.
func Convert_v1alpha1_TanzuKubernetesReleaseSpec_To_v1alpha3_TanzuKubernetesReleaseSpec(in *TanzuKubernetesReleaseSpec, out *v1alpha3.TanzuKubernetesReleaseSpec, s apiconversion.Scope) error {
	out.Kubernetes.Version = in.KubernetesVersion
	out.Kubernetes.ImageRepository = in.Repository

	// Transform the container images.
	for index, image := range in.Images {
		switch in.Images[index].Name {
		case etcdContainerImageName:
			out.Kubernetes.Etcd = &v1alpha3.ContainerImageInfo{
				ImageRepository: image.Repository,
				ImageTag:        image.Tag,
			}
		case pauseContainerImageName:
			out.Kubernetes.Pause = &v1alpha3.ContainerImageInfo{
				ImageRepository: image.Repository,
				ImageTag:        image.Tag,
			}
		case corednsContainerImageName:
			out.Kubernetes.CoreDNS = &v1alpha3.ContainerImageInfo{
				ImageRepository: image.Repository,
				ImageTag:        image.Tag,
			}
		default:
			break
		}
	}

	return autoConvert_v1alpha1_TanzuKubernetesReleaseSpec_To_v1alpha3_TanzuKubernetesReleaseSpec(in, out, s)
}

// Convert_v1alpha3_TanzuKubernetesReleaseSpec_To_v1alpha1_TanzuKubernetesReleaseSpec will convert the compatible types in TanzuKubernetesReleaseSpec v1alpha3 to v1alpha1 equivalent.
// nolint:revive,stylecheck // Generated conversion stubs have underscores in function names.
func Convert_v1alpha3_TanzuKubernetesReleaseSpec_To_v1alpha1_TanzuKubernetesReleaseSpec(in *v1alpha3.TanzuKubernetesReleaseSpec, out *TanzuKubernetesReleaseSpec, s apiconversion.Scope) error {
	out.KubernetesVersion = in.Kubernetes.Version
	out.Repository = in.Kubernetes.ImageRepository

	// Transform the containerimages.
	// Container images are completely restored from the annotations later.
	// This is to handle the scenario when there are no annotations present.
	if in.Kubernetes.Etcd != nil {
		out.Images = append(out.Images, ContainerImage{
			Name:       etcdContainerImageName,
			Repository: in.Kubernetes.Etcd.ImageRepository,
			Tag:        in.Kubernetes.Etcd.ImageTag,
		})
	}
	if in.Kubernetes.CoreDNS != nil {
		out.Images = append(out.Images, ContainerImage{
			Name:       corednsContainerImageName,
			Repository: in.Kubernetes.CoreDNS.ImageRepository,
			Tag:        in.Kubernetes.CoreDNS.ImageTag,
		})
	}
	if in.Kubernetes.Pause != nil {
		out.Images = append(out.Images, ContainerImage{
			Name:       pauseContainerImageName,
			Repository: in.Kubernetes.Pause.ImageRepository,
			Tag:        in.Kubernetes.Pause.ImageTag,
		})
	}
	return autoConvert_v1alpha3_TanzuKubernetesReleaseSpec_To_v1alpha1_TanzuKubernetesReleaseSpec(in, out, s)
}

func (src *TanzuKubernetesReleaseList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.TanzuKubernetesReleaseList)

	return Convert_v1alpha1_TanzuKubernetesReleaseList_To_v1alpha3_TanzuKubernetesReleaseList(src, dst, nil)
}

//nolint:revive,stylecheck // Much easier to read when the receiver is named as spoke/src/dest.
func (dst *TanzuKubernetesReleaseList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.TanzuKubernetesReleaseList)

	return autoConvert_v1alpha3_TanzuKubernetesReleaseList_To_v1alpha1_TanzuKubernetesReleaseList(src, dst, nil)
}
