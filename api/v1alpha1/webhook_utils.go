package v1alpha1

import (
	"errors"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defNeuron = corev1.Container{
	Name: "neuron",
	Env: []corev1.EnvVar{
		{
			Name:  "LOG_CONSOLE",
			Value: "1",
		},
	},
	Ports: []corev1.ContainerPort{
		{
			Name:     "neuron",
			Protocol: corev1.ProtocolTCP,
			// neuron web port is hardcode in source code
			ContainerPort: 7000,
		},
	},
}

var defEKuiper = corev1.Container{
	Name: "eKuiper",
	Env: []corev1.EnvVar{
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
	},
}

func setDefaultLabels(ins EdgeInterface) {
	labels := ins.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[ManagerByKey] = "edge-operator"
	labels[InstanceKey] = ins.GetName()
	labels[ComponentKey] = string(ins.GetComponentType())
	ins.SetLabels(labels)
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

func setContainerPortsFromEnv(container *corev1.Container) {
	envs := container.Env
	for i := range envs {
		if envs[i].Name == "KUIPER__BASIC__RESTPORT" {
			found := false
			index := 0
			for index <= len(container.Ports)-1 {
				if container.Ports[index].Name == "ekuiper" {
					found = true
					break
				}
				index++
			}
			if found {
				container.Ports[index].ContainerPort = intstr.Parse(envs[i].Value).IntVal
				continue
			}
			container.Ports = append(container.Ports, corev1.ContainerPort{
				Name:          "ekuiper",
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

func setDefaultVolume(ins EdgeInterface) {
	vol := ins.GetVolumeClaimTemplate()
	if vol == nil {
		return
	}
	if vol.Name == "" {
		vol.Name = ins.GetResName()
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
		svc.Name = ins.GetResName()
	}

	svc.Namespace = ins.GetNamespace()
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

func setDefaultNeuronProbe(ins EdgeInterface) {
	neuron := ins.GetNeuron()
	if neuron.ReadinessProbe == nil {
		neuron.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: intstr.FromInt(7000),
				},
			},
		}
	}
	if neuron.LivenessProbe == nil {
		neuron.LivenessProbe = &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: intstr.FromInt(7000),
				},
			},
		}
	}
}

func setDefaultEKuiperProbe(ins EdgeInterface) {
	ekuiper := ins.GetEKuiper()
	port := intstr.FromInt(9081)
	for i := range ekuiper.Env {
		if ekuiper.Env[i].Name == "KUIPER__BASIC__RESTPORT" {
			port = intstr.Parse(ekuiper.Env[i].Value)
		}
	}

	if ekuiper.ReadinessProbe == nil {
		ekuiper.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: port,
				},
			},
		}
	}
	if ekuiper.LivenessProbe == nil {
		ekuiper.LivenessProbe = &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			FailureThreshold:    12,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "",
					Port: port,
				},
			},
		}
	}
}

func validateNeuronContainer(ins EdgeInterface) error {
	neuron := ins.GetNeuron()

	if neuron.Name == "" {
		return errors.New("neuron container name is empty")
	}

	if neuron.Image == "" {
		return errors.New("neuron container image is empty")
	}

	return nil
}

func validateEKuiperContainer(ins EdgeInterface) error {
	ekuiper := ins.GetEKuiper()

	if ekuiper.Name == "" {
		return errors.New("ekuiper container name is empty")
	}

	if ekuiper.Image == "" {
		return errors.New("ekuiper container image is empty")
	}

	if !strings.HasSuffix(ekuiper.Image, "-slim-python") && !strings.HasSuffix(ekuiper.Image, "-slim") {
		return errors.New("ekuiper container image must be slim or slim-python")
	}
	return nil
}
