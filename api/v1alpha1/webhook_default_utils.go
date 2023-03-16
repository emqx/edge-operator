package v1alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func setDefaultLabels(ins EdgeInterface) {
	labels := ins.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[ManagedByKey] = "edge-operator"
	labels[InstanceKey] = ins.GetName()
	labels[ComponentKey] = string(ins.GetComponentType())
	ins.SetLabels(labels)
}

func setDefaultNeuronContainer(ins EdgeInterface) {
	neuron := ins.GetNeuron()
	if neuron.Name == "" {
		neuron.Name = "neuron"
	}
	if neuron.ImagePullPolicy == "" {
		neuron.ImagePullPolicy = corev1.PullAlways
		i := strings.Split(ins.GetNeuron().Image, ":")
		if len(i) == 2 && !strings.Contains(i[1], "latest") {
			neuron.ImagePullPolicy = corev1.PullIfNotPresent
		}
	}
	if neuron.TerminationMessagePath == "" {
		neuron.TerminationMessagePath = corev1.TerminationMessagePathDefault
	}
	if neuron.TerminationMessagePolicy == "" {
		neuron.TerminationMessagePolicy = corev1.TerminationMessageReadFile
	}
	neuron.Env = mergeEnv(neuron.Env, []corev1.EnvVar{
		{
			Name:  "LOG_CONSOLE",
			Value: "1",
		},
	})
	neuron.Ports = mergeContainerPorts(neuron.Ports, []corev1.ContainerPort{
		{
			Name:     "neuron",
			Protocol: corev1.ProtocolTCP,
			// neuron web port is hardcode in source code
			ContainerPort: 7000,
		},
	})
	if neuron.ReadinessProbe == nil {
		neuron.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/",
					Port:   intstr.FromInt(7000),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      1,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			FailureThreshold:    12,
		}
	}
	if neuron.LivenessProbe == nil {
		neuron.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/",
					Port:   intstr.FromInt(7000),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      1,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			FailureThreshold:    12,
		}
	}
}

func setDefaultEKuiperContainer(ins EdgeInterface) {
	ekuiper := ins.GetEKuiper()
	if ekuiper.Name == "" {
		ekuiper.Name = "ekuiper"
	}
	if ekuiper.ImagePullPolicy == "" {
		ekuiper.ImagePullPolicy = corev1.PullAlways
		i := strings.Split(ins.GetEKuiper().Image, ":")
		if len(i) == 2 && !strings.Contains(i[1], "latest") {
			ekuiper.ImagePullPolicy = corev1.PullIfNotPresent
		}
	}
	if ekuiper.TerminationMessagePath == "" {
		ekuiper.TerminationMessagePath = corev1.TerminationMessagePathDefault
	}
	if ekuiper.TerminationMessagePolicy == "" {
		ekuiper.TerminationMessagePolicy = corev1.TerminationMessageReadFile
	}
	ekuiper.Env = mergeEnv(ekuiper.Env, []corev1.EnvVar{
		{
			Name:  "KUIPER__BASIC__RESTPORT",
			Value: "9081",
		},
		{
			Name:  "KUIPER__BASIC__IGNORECASE",
			Value: "false",
		},
		{
			Name:  "KUIPER__BASIC__CONSOLELOG",
			Value: "true",
		},
	})
	containerPort := intstr.Parse("9081")
	for _, env := range ekuiper.Env {
		if env.Name == "KUIPER__BASIC__RESTPORT" {
			containerPort = intstr.Parse(env.Value)
		}
	}
	ekuiper.Ports = mergeContainerPorts(ekuiper.Ports, []corev1.ContainerPort{
		{
			Name:          "ekuiper",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: containerPort.IntVal,
		},
	})
	if ekuiper.ReadinessProbe == nil {
		ekuiper.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/",
					Port:   containerPort,
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      1,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			FailureThreshold:    12,
		}
	}
	if ekuiper.LivenessProbe == nil {
		ekuiper.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/",
					Port:   containerPort,
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      1,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			FailureThreshold:    12,
		}
	}
}

func setDefaultVolume(ins EdgeInterface) {
	vol := ins.GetVolumeClaimTemplate()
	if vol == nil {
		return
	}
	if vol.Name == "" {
		vol.Name = ins.GetName()
	}
	vol.Namespace = ins.GetNamespace()
	mergeLabels(vol, ins)
	mergeAnnotations(vol, ins)
}

func setDefaultService(ins EdgeInterface) {
	svc := ins.GetServiceTemplate()
	if svc == nil {
		return
	}

	if svc.Name == "" {
		svc.Name = ins.GetName()
	}

	svc.Namespace = ins.GetNamespace()
	mergeLabels(svc, ins)
	mergeAnnotations(svc, ins)

	svc.Spec.Selector = mergeMap(svc.Spec.Selector, ins.GetLabels())

	var sPorts []corev1.ServicePort
	transPort := func(cPorts []corev1.ContainerPort) {
		for i := range cPorts {
			sPorts = append(sPorts, corev1.ServicePort{
				Name:       cPorts[i].Name,
				Protocol:   cPorts[i].Protocol,
				Port:       cPorts[i].ContainerPort,
				TargetPort: intstr.FromInt(int(cPorts[i].ContainerPort)),
			})
		}
	}

	eKuiper := ins.GetEKuiper()
	if eKuiper != nil {
		transPort(eKuiper.Ports)
	}

	neuron := ins.GetNeuron()
	if neuron != nil {
		transPort(neuron.Ports)
	}

	mergeServicePort(svc, sPorts)
}

// mergeLabels merges the labels specified by the operator into
// on object's metadata.
//
// This will return whether the target's labels have changed.
func mergeLabels(target, desired metav1.Object) {
	target.SetLabels(mergeMap(target.GetLabels(), desired.GetLabels()))
}

// mergeAnnotations merges the annotations specified by the operator into
// on object's metadata.
//
// This will return whether the target's annotations have changed.
func mergeAnnotations(target, desired metav1.Object) {
	desiredAnnotations := desired.GetAnnotations()
	if desiredAnnotations != nil {
		delete(desiredAnnotations, corev1.LastAppliedConfigAnnotation)
	}
	target.SetAnnotations(mergeMap(target.GetAnnotations(), desiredAnnotations))
}

// mergeMap merges a map into another map.
//
// This will return whether the target's values have changed.
func mergeMap(target map[string]string, desired map[string]string) map[string]string {
	if target == nil {
		target = make(map[string]string)
	}
	for key, value := range desired {
		if target[key] != value {
			target[key] = value
		}
	}
	return target
}

// mergeEnv adds environment variables to an existing environment, unless
// environment variables with the same name are already present.
func mergeEnv(target, desired []corev1.EnvVar) []corev1.EnvVar {
	envs := append(target, desired...)
	result := make([]corev1.EnvVar, 0, len(envs))
	temp := map[string]struct{}{}

	for _, item := range envs {
		if _, ok := temp[item.Name]; !ok {
			temp[item.Name] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// mergeContainerPorts merge the same name and containerPort's port
func mergeContainerPorts(target, desired []corev1.ContainerPort) []corev1.ContainerPort {
	ports := append(target, desired...)
	result := make([]corev1.ContainerPort, 0, len(ports))
	temp := map[string]struct{}{}

	for _, item := range ports {
		if _, ok := temp[item.Name]; !ok {
			temp[item.Name] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func mergeServicePort(svc *corev1.Service, required []corev1.ServicePort) {
	ports := append(svc.Spec.Ports, required...)
	result := make([]corev1.ServicePort, 0, len(ports))
	temp := map[string]struct{}{}

	for _, item := range ports {
		if _, ok := temp[item.Name]; !ok {
			temp[item.Name] = struct{}{}
			result = append(result, item)
		}
	}
	svc.Spec.Ports = result
}
