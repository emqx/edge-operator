package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	eKuiperBasePort     = "KUIPER__BASIC__PORT"
	eKuiperBaseRestPort = "KUIPER__BASIC__RESTPORT"
)

var defEKuiper = corev1.Container{
	Name: "eKuiper",
	Env: []corev1.EnvVar{
		{
			Name:  eKuiperBasePort,
			Value: "20498",
		},
		{
			Name:  eKuiperBaseRestPort,
			Value: "9081",
		},
	},
}

var defNeuron = corev1.Container{
	Name: "neuron",
}

func getCRObjectMeta(insName string, compType ComponentType) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels: map[string]string{
			ManagerByKey: "edge-operator",
			InstanceKey:  insName,
			ComponentKey: compType.String(),
		},
	}
}

// extendEnv adds environment variables to an existing environment, unless
// environment variables with the same name are already present.
func extendEnv(container *corev1.Container, env []corev1.EnvVar) {
	existingVars := make(map[string]bool, len(container.Env))

	for _, envVar := range container.Env {
		existingVars[envVar.Name] = true
	}

	for _, envVar := range env {
		if !existingVars[envVar.Name] {
			container.Env = append(container.Env, envVar)
		}
	}
}

// mergeLabels merges the labels specified by the operator into
// on object's metadata.
//
// This will return whether the target's labels have changed.
func mergeLabelsInMetadata(target *metav1.ObjectMeta, desired metav1.ObjectMeta) bool {
	if target.Labels == nil {
		target.Labels = make(map[string]string)
	}
	return mergeMap(target.Labels, desired.Labels)
}

// mergeAnnotations merges the annotations specified by the operator into
// on object's metadata.
//
// This will return whether the target's annotations have changed.
func mergeAnnotations(target *metav1.ObjectMeta, desired metav1.ObjectMeta) bool {
	if target.Annotations == nil {
		target.Annotations = make(map[string]string)
	}
	delete(desired.Annotations, corev1.LastAppliedConfigAnnotation)
	return mergeMap(target.Annotations, desired.Annotations)
}

// mergeMap merges a map into another map.
//
// This will return whether the target's values have changed.
func mergeMap(target map[string]string, desired map[string]string) bool {
	changed := false
	for key, value := range desired {
		if target[key] != value {
			target[key] = value
			changed = true
		}
	}
	return changed
}

// mergePorts merge the same name of env to user defined port
func mergePorts(container *corev1.Container, envPorts map[string]string) {
	ports := container.Ports
	for name, port := range envPorts {
		found := false
		for i := range ports {
			if ports[i].Name == name {
				found = true
				break
			}
		}

		if !found {
			ports = append(ports, corev1.ContainerPort{
				Name:          name,
				ContainerPort: intstr.Parse(port).IntVal,
				Protocol:      corev1.ProtocolTCP,
			})
		}
	}
	container.Ports = ports
}

func setDefaultEKuiperPort(target *corev1.Container, env []corev1.EnvVar) {
	extendEnv(target, env)

	envPorts := make(map[string]string)
	for i := range target.Env {
		value := target.Env[i].Value
		switch target.Env[i].Name {
		case eKuiperBasePort:
			envPorts["port"] = value
		case eKuiperBaseRestPort:
			envPorts["rest-port"] = value
		}
	}
	mergePorts(target, envPorts)
}

func setDefaultService(ins EdgeInterface) {
	svc := ins.GetServiceTemplate()
	if svc == nil {
		return
	}

	if svc.Name == "" {
		svc.Name = ins.GetComponentType().GetResName(ins)
	}

	if len(svc.Spec.Selector) == 0 {
		svc.Spec.Selector = make(map[string]string)
	}
	mergeMap(svc.Spec.Selector, ins.GetLabels())

	instanceMeta := *ins.GetObjectMeta().(*metav1.ObjectMeta)
	mergeAnnotations(&svc.ObjectMeta, instanceMeta)
	mergeLabelsInMetadata(&svc.ObjectMeta, instanceMeta)

	var sPorts []corev1.ServicePort
	appendPort := func(cPorts []corev1.ContainerPort) {
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
		appendPort(eKuiper.Ports)
	}

	neuron := ins.GetNeuron()
	if neuron != nil {
		appendPort(neuron.Ports)
	}

	mergeServicePort(svc, sPorts)
}

func mergeServicePort(svc *corev1.Service, required []corev1.ServicePort) {
	target := svc.Spec.Ports
	for i := range required {
		found := false
		for j := range target {
			if target[j].Name == required[i].Name {
				found = true
				break
			}
		}

		if !found {
			target = append(target, *required[i].DeepCopy())
		}
	}
	svc.Spec.Ports = target
}

func setDefaultVolume(ins EdgeInterface) {
	vol := ins.GetVolumeClaimTemplate()
	if vol == nil {
		return
	}

	instanceMeta := *ins.GetObjectMeta().(*metav1.ObjectMeta)
	mergeAnnotations(&vol.ObjectMeta, instanceMeta)
	mergeLabelsInMetadata(&vol.ObjectMeta, instanceMeta)
}
