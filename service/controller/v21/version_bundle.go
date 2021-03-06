package v21

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "calico",
				Description: "Update calico to 3.6.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Pin calico-kube-controllers to master.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "containerlinux",
				Description: "Update to 2023.5.0.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Update kubernetes to 1.14.1. More info here: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG-1.14.md",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Tolerate all taints for calico and kube-proxy daemon sets.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "kvm-operator",
				Description: "Use XFS for docker and kubelet volumes.",
				Kind:        versionbundle.KindFixed,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.6.1",
			},
			{
				Name:    "containerlinux",
				Version: "2023.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.12",
			},
			{
				Name:    "kubernetes",
				Version: "1.14.1",
			},
		},
		Name:    "kvm-operator",
		Version: "3.5.0",
	}
}
