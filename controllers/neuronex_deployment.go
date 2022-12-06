package controllers

import (
	"context"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type childSubReconciler interface {
	reconcile(ctx context.Context, r *NeuronEXReconciler, instance *edgev1alpha1.NeuronEX) *requeue
	updateDeployment(deploy *appsv1.Deployment, instance *edgev1alpha1.NeuronEX)
}

type neuronEXDeploy struct {
	subReconcilerList []childSubReconciler
}

func newNeuronEXDeploy() neuronEXDeploy {
	return neuronEXDeploy{
		subReconcilerList: []childSubReconciler{
			addPVC{},
			ekuiperTool{},
		},
	}
}

func (sub neuronEXDeploy) reconcile(ctx context.Context, r *NeuronEXReconciler, instance *edgev1alpha1.NeuronEX) *requeue {
	deploy := sub.getDeployment(instance)

	for _, subReconciler := range sub.subReconcilerList {
		if err := subReconciler.reconcile(ctx, r, instance); err != nil {
			return err
		}
		subReconciler.updateDeployment(deploy, instance)
	}

	if err := createOrUpdate(ctx, r, instance, deploy); err != nil {
		return &requeue{curError: emperror.Wrap(err, "failed to create or update deployment")}
	}

	return nil
}

func (sub neuronEXDeploy) getDeployment(instance *edgev1alpha1.NeuronEX) *appsv1.Deployment {
	labels := instance.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	delete(labels, "kubectl.kubernetes.io/last-applied-configuration")

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Annotations: instance.Annotations,
			Labels:      labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: instance.Annotations,
				},
				Spec: sub.getPodSpec(instance),
			},
		},
	}
	deploy.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))
	return deploy
}

func (sub neuronEXDeploy) getPodSpec(instance *edgev1alpha1.NeuronEX) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			*sub.getNeuronContainer(instance.Spec.Neuron.DeepCopy()),
			*sub.getEkuiperContainer(instance.Spec.EKuiper.DeepCopy()),
		},
		Volumes:                       instance.Spec.Volumes,
		InitContainers:                instance.Spec.InitContainers,
		EphemeralContainers:           instance.Spec.EphemeralContainers,
		RestartPolicy:                 instance.Spec.RestartPolicy,
		TerminationGracePeriodSeconds: instance.Spec.TerminationGracePeriodSeconds,
		ActiveDeadlineSeconds:         instance.Spec.ActiveDeadlineSeconds,
		DNSPolicy:                     instance.Spec.DNSPolicy,
		NodeSelector:                  instance.Spec.NodeSelector,
		ServiceAccountName:            instance.Spec.ServiceAccountName,
		DeprecatedServiceAccount:      instance.Spec.DeprecatedServiceAccount,
		AutomountServiceAccountToken:  instance.Spec.AutomountServiceAccountToken,
		NodeName:                      instance.Spec.NodeName,
		HostNetwork:                   instance.Spec.HostNetwork,
		HostPID:                       instance.Spec.HostPID,
		HostIPC:                       instance.Spec.HostIPC,
		ShareProcessNamespace:         instance.Spec.ShareProcessNamespace,
		SecurityContext:               instance.Spec.PodSecurityContext,
		ImagePullSecrets:              instance.Spec.ImagePullSecrets,
		Hostname:                      instance.Spec.Hostname,
		Subdomain:                     instance.Spec.Subdomain,
		Affinity:                      instance.Spec.Affinity,
		SchedulerName:                 instance.Spec.SchedulerName,
		Tolerations:                   instance.Spec.Tolerations,
		HostAliases:                   instance.Spec.HostAliases,
		PriorityClassName:             instance.Spec.PriorityClassName,
		Priority:                      instance.Spec.Priority,
		DNSConfig:                     instance.Spec.DNSConfig,
		ReadinessGates:                instance.Spec.ReadinessGates,
		RuntimeClassName:              instance.Spec.RuntimeClassName,
		EnableServiceLinks:            instance.Spec.EnableServiceLinks,
		PreemptionPolicy:              instance.Spec.PreemptionPolicy,
		Overhead:                      instance.Spec.Overhead,
		TopologySpreadConstraints:     instance.Spec.TopologySpreadConstraints,
		SetHostnameAsFQDN:             instance.Spec.SetHostnameAsFQDN,
		OS:                            instance.Spec.OS,
		HostUsers:                     instance.Spec.HostUsers,
	}
}

func (sub neuronEXDeploy) getNeuronContainer(neuron *corev1.Container) *corev1.Container {
	return neuron
}

func (sub neuronEXDeploy) getEkuiperContainer(ekuiper *corev1.Container) *corev1.Container {
	ekuiper.Env = append([]corev1.EnvVar{
		{
			Name:  "MQTT_SOURCE__DEFAULT__SERVER",
			Value: "tcp://broker.emqx.io:1883",
		},
		{
			Name:  "KUIPER__BASIC__FILELOG",
			Value: "false",
		},
		{
			Name:  "KUIPER__BASIC__CONSOLELOG",
			Value: "true",
		},
	}, ekuiper.Env...)

	return ekuiper
}
