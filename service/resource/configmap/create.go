package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToCreate, err := toConfigMaps(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	// Create the config maps in the Kubernetes API.
	if len(configMapsToCreate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the config maps in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, configMap := range configMapsToCreate {
			_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(configMap)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the config maps in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the config maps do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps have to be created")

	var configMapsToCreate []*apiv1.ConfigMap

	for _, desiredConfigMap := range desiredConfigMaps {
		if !containsConfigMap(currentConfigMaps, desiredConfigMap) {
			configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps that have to be created", len(configMapsToCreate)))

	return configMapsToCreate, nil
}