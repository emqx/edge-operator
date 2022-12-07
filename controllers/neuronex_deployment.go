package controllers

import (
	"context"

	emperror "emperror.dev/errors"
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deploySubReconciler interface {
	reconcile(ctx context.Context, r *NeuronEXReconciler, instance edgev1alpha1.EdgeInterface) *requeue
	updateDeployment(deploy *appsv1.Deployment, instance edgev1alpha1.EdgeInterface)
}

type neuronEXDeploy struct {
	subReconcilerList []deploySubReconciler
}

func newNeuronEXDeploy() neuronEXDeploy {
	return neuronEXDeploy{
		subReconcilerList: []deploySubReconciler{
			addPVC{},
			ekuiperTool{},
		},
	}
}

func (sub neuronEXDeploy) reconcile(ctx context.Context, r *NeuronEXReconciler, instance edgev1alpha1.EdgeInterface) *requeue {
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

func (sub neuronEXDeploy) getDeployment(instance edgev1alpha1.EdgeInterface) *appsv1.Deployment {
	labels := instance.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	delete(labels, "kubectl.kubernetes.io/last-applied-configuration")

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.GetName(),
			Namespace:   instance.GetNamespace(),
			Annotations: instance.GetAnnotations(),
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
					Annotations: instance.GetAnnotations(),
				},
				Spec: sub.getPodSpec(instance),
			},
		},
	}
	deploy.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))
	return deploy
}

func (sub neuronEXDeploy) getPodSpec(instance edgev1alpha1.EdgeInterface) corev1.PodSpec {
	containers := []corev1.Container{}
	if instance.GetNeuron() != nil {
		containers = append(containers, *sub.getNeuronContainer(instance.GetNeuron().DeepCopy()))
	}
	if instance.GetEKuiper() != nil {
		containers = append(containers, *sub.getEkuiperContainer(instance.GetEKuiper().DeepCopy()))
	}

	return corev1.PodSpec{
		Containers:                    containers,
		Volumes:                       instance.GetEdgePodSpec().Volumes,
		InitContainers:                instance.GetEdgePodSpec().InitContainers,
		EphemeralContainers:           instance.GetEdgePodSpec().EphemeralContainers,
		RestartPolicy:                 instance.GetEdgePodSpec().RestartPolicy,
		TerminationGracePeriodSeconds: instance.GetEdgePodSpec().TerminationGracePeriodSeconds,
		ActiveDeadlineSeconds:         instance.GetEdgePodSpec().ActiveDeadlineSeconds,
		DNSPolicy:                     instance.GetEdgePodSpec().DNSPolicy,
		NodeSelector:                  instance.GetEdgePodSpec().NodeSelector,
		ServiceAccountName:            instance.GetEdgePodSpec().ServiceAccountName,
		DeprecatedServiceAccount:      instance.GetEdgePodSpec().DeprecatedServiceAccount,
		AutomountServiceAccountToken:  instance.GetEdgePodSpec().AutomountServiceAccountToken,
		NodeName:                      instance.GetEdgePodSpec().NodeName,
		HostNetwork:                   instance.GetEdgePodSpec().HostNetwork,
		HostPID:                       instance.GetEdgePodSpec().HostPID,
		HostIPC:                       instance.GetEdgePodSpec().HostIPC,
		ShareProcessNamespace:         instance.GetEdgePodSpec().ShareProcessNamespace,
		SecurityContext:               instance.GetEdgePodSpec().PodSecurityContext,
		ImagePullSecrets:              instance.GetEdgePodSpec().ImagePullSecrets,
		Hostname:                      instance.GetEdgePodSpec().Hostname,
		Subdomain:                     instance.GetEdgePodSpec().Subdomain,
		Affinity:                      instance.GetEdgePodSpec().Affinity,
		SchedulerName:                 instance.GetEdgePodSpec().SchedulerName,
		Tolerations:                   instance.GetEdgePodSpec().Tolerations,
		HostAliases:                   instance.GetEdgePodSpec().HostAliases,
		PriorityClassName:             instance.GetEdgePodSpec().PriorityClassName,
		Priority:                      instance.GetEdgePodSpec().Priority,
		DNSConfig:                     instance.GetEdgePodSpec().DNSConfig,
		ReadinessGates:                instance.GetEdgePodSpec().ReadinessGates,
		RuntimeClassName:              instance.GetEdgePodSpec().RuntimeClassName,
		EnableServiceLinks:            instance.GetEdgePodSpec().EnableServiceLinks,
		PreemptionPolicy:              instance.GetEdgePodSpec().PreemptionPolicy,
		Overhead:                      instance.GetEdgePodSpec().Overhead,
		TopologySpreadConstraints:     instance.GetEdgePodSpec().TopologySpreadConstraints,
		SetHostnameAsFQDN:             instance.GetEdgePodSpec().SetHostnameAsFQDN,
		OS:                            instance.GetEdgePodSpec().OS,
		HostUsers:                     instance.GetEdgePodSpec().HostUsers,
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
