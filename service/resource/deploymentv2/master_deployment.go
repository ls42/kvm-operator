package deploymentv2

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func newMasterDeployments(customObject v1alpha1.KVMConfig) ([]*extensionsv1.Deployment, error) {
	var deployments []*extensionsv1.Deployment

	privileged := true
	replicas := int32(1)

	for i, masterNode := range customObject.Spec.Cluster.Masters {
		capabilities := customObject.Spec.KVM.Masters[i]

		cpuQuantity, err := keyv2.CPUQuantity(capabilities)
		if err != nil {
			return nil, microerror.Maskf(err, "creating CPU quantity")
		}

		memoryQuantity, err := keyv2.MemoryQuantity(capabilities)
		if err != nil {
			return nil, microerror.Maskf(err, "creating memory quantity")
		}

		storageType := keyv2.StorageType(customObject)

		// During migration, some TPOs do not have storage type set.
		// This specifies a default, until all TPOs have the correct storage type set.
		// tl;dr - this shouldn't be here. If all TPOs have storageType, remove it.
		if storageType == "" {
			storageType = "hostPath"
		}

		var etcdVolume apiv1.Volume
		if storageType == "hostPath" {
			etcdVolume = apiv1.Volume{
				Name: "etcd-data",
				VolumeSource: apiv1.VolumeSource{
					HostPath: &apiv1.HostPathVolumeSource{
						Path: keyv2.MasterHostPathVolumeDir(keyv2.ClusterID(customObject), keyv2.VMNumber(i)),
					},
				},
			}
		} else if storageType == "persistentVolume" {
			etcdVolume = apiv1.Volume{
				Name: "etcd-data",
				VolumeSource: apiv1.VolumeSource{
					PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
						ClaimName: keyv2.EtcdPVCName(keyv2.ClusterID(customObject), keyv2.VMNumber(i)),
					},
				},
			}
		} else {
			return nil, microerror.Maskf(wrongTypeError, "unknown storageType: '%s'", keyv2.StorageType(customObject))
		}
		deployment := &extensionsv1.Deployment{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "deployment",
				APIVersion: "extensions/v1beta",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: keyv2.DeploymentName(keyv2.MasterID, masterNode.ID),
				Annotations: map[string]string{
					VersionBundleVersionAnnotation: keyv2.VersionBundleVersion(customObject),
				},
				Labels: map[string]string{
					"cluster":  keyv2.ClusterID(customObject),
					"customer": keyv2.ClusterCustomer(customObject),
					"app":      keyv2.MasterID,
					"node":     masterNode.ID,
				},
			},
			Spec: extensionsv1.DeploymentSpec{
				Strategy: extensionsv1.DeploymentStrategy{
					Type: extensionsv1.RecreateDeploymentStrategyType,
				},
				Replicas: &replicas,
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: apismetav1.ObjectMeta{
						GenerateName: keyv2.MasterID,
						Labels: map[string]string{
							"app":      keyv2.MasterID,
							"cluster":  keyv2.ClusterID(customObject),
							"customer": keyv2.ClusterCustomer(customObject),
							"node":     masterNode.ID,
						},
						Annotations: map[string]string{},
					},
					Spec: apiv1.PodSpec{
						Affinity:    newMasterPodAfinity(customObject),
						HostNetwork: true,
						NodeSelector: map[string]string{
							"role": keyv2.MasterID,
						},
						Volumes: []apiv1.Volume{
							{
								Name: "cloud-config",
								VolumeSource: apiv1.VolumeSource{
									ConfigMap: &apiv1.ConfigMapVolumeSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: keyv2.ConfigMapName(customObject, masterNode, keyv2.MasterID),
										},
									},
								},
							},
							etcdVolume,
							{
								Name: "images",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: "/home/core/images/",
									},
								},
							},
							{
								Name: "rootfs",
								VolumeSource: apiv1.VolumeSource{
									EmptyDir: &apiv1.EmptyDirVolumeSource{},
								},
							},
							{
								Name: "flannel",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: keyv2.FlannelEnvPathPrefix,
									},
								},
							},
						},
						Containers: []apiv1.Container{
							{
								Name:            "k8s-endpoint-updater",
								Image:           customObject.Spec.KVM.EndpointUpdater.Docker.Image,
								ImagePullPolicy: apiv1.PullIfNotPresent,
								Command: []string{
									"/bin/sh",
									"-c",
									"/opt/k8s-endpoint-updater update --provider.bridge.name=" + keyv2.NetworkBridgeName(customObject) +
										" --service.kubernetes.cluster.namespace=" + keyv2.ClusterNamespace(customObject) +
										" --service.kubernetes.cluster.service=" + keyv2.MasterID +
										" --service.kubernetes.inCluster=true" +
										" --service.kubernetes.pod.name=${POD_NAME}",
								},
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								Env: []apiv1.EnvVar{
									{
										Name: "POD_NAME",
										ValueFrom: &apiv1.EnvVarSource{
											FieldRef: &apiv1.ObjectFieldSelector{
												APIVersion: "v1",
												FieldPath:  "metadata.name",
											},
										},
									},
								},
							},
							{
								Name:            "k8s-kvm",
								Image:           customObject.Spec.KVM.K8sKVM.Docker.Image,
								ImagePullPolicy: apiv1.PullIfNotPresent,
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								Args: []string{
									keyv2.MasterID,
								},
								Env: []apiv1.EnvVar{
									{
										Name:  "CORES",
										Value: fmt.Sprintf("%d", capabilities.CPUs),
									},
									{
										Name:  "DISK",
										Value: fmt.Sprintf("%.0fG", capabilities.Disk),
									},
									{
										Name: "HOSTNAME",
										ValueFrom: &apiv1.EnvVarSource{
											FieldRef: &apiv1.ObjectFieldSelector{
												APIVersion: "v1",
												FieldPath:  "metadata.name",
											},
										},
									},
									{
										Name:  "NETWORK_BRIDGE_NAME",
										Value: keyv2.NetworkBridgeName(customObject),
									},
									{
										Name:  "NETWORK_TAP_NAME",
										Value: keyv2.NetworkTapName(customObject),
									},
									{
										Name: "MEMORY",
										// TODO provide memory like disk as float64 and format here.
										Value: capabilities.Memory,
									},
									{
										Name:  "ROLE",
										Value: keyv2.MasterID,
									},
									{
										Name:  "CLOUD_CONFIG_PATH",
										Value: "/cloudconfig/user_data",
									},
								},
								LivenessProbe: &apiv1.Probe{
									InitialDelaySeconds: keyv2.InitialDelaySeconds,
									TimeoutSeconds:      keyv2.TimeoutSeconds,
									PeriodSeconds:       keyv2.PeriodSeconds,
									FailureThreshold:    keyv2.FailureThreshold,
									SuccessThreshold:    keyv2.SuccessThreshold,
									Handler: apiv1.Handler{
										HTTPGet: &apiv1.HTTPGetAction{
											Path: keyv2.HealthEndpoint,
											Port: intstr.IntOrString{IntVal: keyv2.LivenessPort(customObject)},
											Host: keyv2.ProbeHost,
										},
									},
								},
								Resources: apiv1.ResourceRequirements{
									Requests: apiv1.ResourceList{
										apiv1.ResourceCPU:    cpuQuantity,
										apiv1.ResourceMemory: memoryQuantity,
									},
								},
								VolumeMounts: []apiv1.VolumeMount{
									{
										Name:      "cloud-config",
										MountPath: "/cloudconfig/",
									},
									{
										Name:      "etcd-data",
										MountPath: "/etc/kubernetes/data/etcd/",
									},
									{
										Name:      "images",
										MountPath: "/usr/code/images/",
									},
									{
										Name:      "rootfs",
										MountPath: "/usr/code/rootfs/",
									},
								},
							},
							{
								Name:            "k8s-kvm-health",
								Image:           keyv2.K8SKVMHealthDocker,
								ImagePullPolicy: apiv1.PullAlways,
								Env: []apiv1.EnvVar{
									{
										Name:  "LISTEN_ADDRESS",
										Value: keyv2.HealthListenAddress(customObject),
									},
									{
										Name:  "NETWORK_ENV_FILE_PATH",
										Value: keyv2.NetworkEnvFilePath(customObject),
									},
								},
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								VolumeMounts: []apiv1.VolumeMount{
									{
										Name:      "flannel",
										MountPath: keyv2.FlannelEnvPathPrefix,
									},
								},
							},
						},
					},
				},
			},
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}