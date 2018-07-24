package endpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	endpointToDelete, err := toK8sEndpoint(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	// The endpoint resource is reconciled by watching pods. Pods get deleted at
	// times. We do not want to delete the whole endpoint only because one pod is
	// gone. We only delete the whole endpoint when it does not contain any IP
	// anymore. Removing IPs is done on update events.
	if endpointToDelete != nil && isEmptyEndpoint(*endpointToDelete) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting endpoint '%s'", endpointToDelete.GetName()))

		err = r.k8sClient.CoreV1().Endpoints(endpointToDelete.Namespace).Delete(endpointToDelete.Name, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted endpoint '%s'", endpointToDelete.GetName()))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting endpoint '%s'", endpointToDelete.GetName()))
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	deleteChange, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	updateChange, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(deleteChange)
	patch.SetUpdateChange(updateChange)

	return patch, nil
}

func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	ips := cutIPs(currentEndpoint.IPs, desiredEndpoint.IPs)

	if currentEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if desiredEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if len(ips) > 0 {
		return nil, nil
	}

	endpoint := &Endpoint{
		ServiceName:      currentEndpoint.ServiceName,
		ServiceNamespace: currentEndpoint.ServiceNamespace,
		IPs:              ips,
	}
	deleteChange, err := r.newK8sEndpoint(endpoint)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return deleteChange, nil
}

func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	ips := cutIPs(currentEndpoint.IPs, desiredEndpoint.IPs)

	if currentEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if desiredEndpoint == nil {
		return nil, nil // Nothing to do.
	}
	if len(ips) == 0 {
		return nil, nil
	}

	endpoint := &Endpoint{
		ServiceName:      currentEndpoint.ServiceName,
		ServiceNamespace: currentEndpoint.ServiceNamespace,
		IPs:              ips,
	}
	updateChange, err := r.newK8sEndpoint(endpoint)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return updateChange, nil
}
