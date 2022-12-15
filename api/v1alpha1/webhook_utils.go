package v1alpha1

import (
	"strings"

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
	Ports: []corev1.ContainerPort{
		{
			Name:     "web",
			Protocol: corev1.ProtocolTCP,
			// neuron web port is hardcode in source code
			ContainerPort: 7000,
		},
	},
}

func getCRObjectMeta(insName string, compType ComponentType) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels: map[string]string{
			ManagerByKey: "edge-operator",
			InstanceKey:  insName,
			ComponentKey: string(compType),
		},
	}
}

// mergeEnv adds environment variables to an existing environment, unless
// environment variables with the same name are already present.
func mergeEnv(target, desired *corev1.Container) {
	existingVars := make(map[string]struct{}, len(target.Env))

	for _, envVar := range target.Env {
		existingVars[envVar.Name] = struct{}{}
	}

	for _, envVar := range desired.Env {
		if _, ok := existingVars[envVar.Name]; !ok {
			target.Env = append(target.Env, envVar)
		}
	}
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
	for key, value := range desired {
		if target[key] != value {
			target[key] = value
		}
	}
	return target
}

func setContainerPortsFromEnv(container *corev1.Container) {
	envs := container.Env
	for i := range envs {
		if envs[i].Name == eKuiperBasePort || envs[i].Name == eKuiperBaseRestPort {
			s := strings.SplitAfter(envs[i].Name, "KUIPER__")[1]
			name := strings.ToLower(strings.ReplaceAll(s, "__", "-"))
			container.Ports = append(container.Ports, corev1.ContainerPort{
				Name:          name,
				ContainerPort: intstr.Parse(envs[i].Value).IntVal,
				Protocol:      corev1.ProtocolTCP,
			})
		}
	}
}

// mergeContainerPorts merge the same name and containerPort's port
func mergeContainerPorts(target, desired *corev1.Container) {
	for _, dPort := range desired.Ports {
		found := false
		for _, tPort := range target.Ports {
			if tPort.Name == dPort.Name || tPort.ContainerPort == dPort.ContainerPort {
				found = true
				break
			}
		}

		if !found {
			target.Ports = append(target.Ports, dPort)
		}
	}
}

func setDefaultService(ins EdgeInterface) {
	svc := ins.GetServiceTemplate()
	if svc == nil {
		return
	}

	if svc.Name == "" {
		svc.Name = ins.GetResName()
	}
	mergeLabels(svc, ins)
	mergeAnnotations(svc, ins)

	svc.Spec.Selector = mergeMap(svc.Spec.Selector, ins.GetLabels())

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

	mergeLabels(vol, ins)
	mergeAnnotations(vol, ins)
}
