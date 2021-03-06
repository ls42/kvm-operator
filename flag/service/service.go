package service

import (
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"

	"github.com/giantswarm/kvm-operator/flag/service/crd"
	"github.com/giantswarm/kvm-operator/flag/service/installation"
	"github.com/giantswarm/kvm-operator/flag/service/tenant"
)

type Service struct {
	CRD          crd.CRD
	Installation installation.Installation
	Kubernetes   kubernetes.Kubernetes
	Tenant       tenant.Tenant
}
