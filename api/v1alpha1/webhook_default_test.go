package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDefault(t *testing.T) {
	for _, ins := range []EdgeInterface{
		&NeuronEX{
			ObjectMeta: metav1.ObjectMeta{
				Name: "neuronex",
			},
			Spec: NeuronEXSpec{
				Neuron: corev1.Container{
					Name: "neuron",
				},
				EKuiper: corev1.Container{
					Name: "ekuiper",
				},
				ServiceTemplate:     &corev1.Service{},
				VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{},
			},
		},
		&Neuron{
			ObjectMeta: metav1.ObjectMeta{
				Name: "neuron",
			},
			Spec: NeuronSpec{
				Neuron: corev1.Container{
					Name: "neuron",
				},
				ServiceTemplate:     &corev1.Service{},
				VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{},
			},
		},
		&EKuiper{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ekuiper",
			},
			Spec: EKuiperSpec{
				EKuiper: corev1.Container{
					Name: "ekuiper",
				},
				ServiceTemplate:     &corev1.Service{},
				VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{},
			},
		},
	} {
		t.Run("check labels", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			got.SetLabels(map[string]string{"foo": "bar"})
			got.Default()
			assert.Equal(t, map[string]string{
				"foo":        "bar",
				ManagerByKey: "edge-operator",
				InstanceKey:  got.GetName(),
				ComponentKey: string(got.GetComponentType()),
			}, got.GetLabels())
		})
		t.Run("check neuron container env", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetNeuron() != nil {
				got.GetNeuron().Env = []corev1.EnvVar{{Name: "foo", Value: "bar"}}
				got.Default()
				assert.ElementsMatch(t, []corev1.EnvVar{
					{Name: "foo", Value: "bar"},
					{Name: "LOG_CONSOLE", Value: "1"},
				}, got.GetNeuron().Env)

				got.GetNeuron().Env = []corev1.EnvVar{{Name: "LOG_CONSOLE", Value: "2"}}
				got.Default()
				assert.ElementsMatch(t, []corev1.EnvVar{
					{Name: "LOG_CONSOLE", Value: "2"},
				}, got.GetNeuron().Env)
			}
		})
		t.Run("check neuron container ports", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetNeuron() != nil {
				got.GetNeuron().Ports = []corev1.ContainerPort{{Name: "foo", ContainerPort: 1234}}
				got.Default()
				assert.ElementsMatch(t, []corev1.ContainerPort{
					{Name: "foo", ContainerPort: 1234},
					{Name: "neuron", Protocol: corev1.ProtocolTCP, ContainerPort: 7000},
				}, got.GetNeuron().Ports)

				got.GetNeuron().Ports = []corev1.ContainerPort{{Name: "neuron", ContainerPort: 1234}}
				got.Default()
				assert.ElementsMatch(t, []corev1.ContainerPort{
					{Name: "neuron", ContainerPort: 1234},
				}, got.GetNeuron().Ports)
			}
		})
		t.Run("check neuron container readiness probe", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetNeuron() != nil {
				got.Default()
				assert.Equal(t, &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "",
							Port:   intstr.FromInt(7000),
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 10,
					TimeoutSeconds:      1,
					PeriodSeconds:       5,
					SuccessThreshold:    1,
					FailureThreshold:    12,
				}, got.GetNeuron().ReadinessProbe)
			}
		})
		t.Run("check neuron container liveness probe", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetNeuron() != nil {
				got.Default()
				assert.Equal(t, &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "",
							Port:   intstr.FromInt(7000),
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 10,
					TimeoutSeconds:      1,
					PeriodSeconds:       5,
					SuccessThreshold:    1,
					FailureThreshold:    12,
				}, got.GetNeuron().LivenessProbe)
			}
		})
		t.Run("check ekuiper container env", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetEKuiper() != nil {
				got.GetEKuiper().Env = []corev1.EnvVar{{Name: "foo", Value: "bar"}}
				got.Default()
				assert.ElementsMatch(t, []corev1.EnvVar{
					{Name: "foo", Value: "bar"},
					{Name: "KUIPER__BASIC__RESTPORT", Value: "9081"},
					{Name: "KUIPER__BASIC__IGNORECASE", Value: "false"},
					{Name: "KUIPER__BASIC__CONSOLELOG", Value: "true"},
				}, got.GetEKuiper().Env)

				got.GetEKuiper().Env = []corev1.EnvVar{{Name: "KUIPER__BASIC__RESTPORT", Value: "9082"}}
				got.Default()
				assert.ElementsMatch(t, []corev1.EnvVar{
					{Name: "KUIPER__BASIC__RESTPORT", Value: "9082"},
					{Name: "KUIPER__BASIC__IGNORECASE", Value: "false"},
					{Name: "KUIPER__BASIC__CONSOLELOG", Value: "true"},
				}, got.GetEKuiper().Env)
			}
		})
		t.Run("check ekuiper container ports", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetEKuiper() != nil {
				got.GetEKuiper().Env = []corev1.EnvVar{}
				got.GetEKuiper().Ports = []corev1.ContainerPort{{Name: "foo", ContainerPort: 1234}}
				got.Default()
				assert.ElementsMatch(t, []corev1.ContainerPort{
					{Name: "foo", ContainerPort: 1234},
					{Name: "ekuiper", Protocol: corev1.ProtocolTCP, ContainerPort: 9081},
				}, got.GetEKuiper().Ports)

				got.GetEKuiper().Env = []corev1.EnvVar{{Name: "KUIPER__BASIC__RESTPORT", Value: "9082"}}
				got.GetEKuiper().Ports = []corev1.ContainerPort{}
				got.Default()
				assert.ElementsMatch(t, []corev1.ContainerPort{
					{Name: "ekuiper", Protocol: corev1.ProtocolTCP, ContainerPort: 9082},
				}, got.GetEKuiper().Ports)
			}
		})
		t.Run("check ekuiper container readiness probe", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetEKuiper() != nil {
				got.GetEKuiper().Env = []corev1.EnvVar{}
				got.Default()
				assert.Equal(t, &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "",
							Port:   intstr.FromInt(9081),
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 10,
					TimeoutSeconds:      1,
					PeriodSeconds:       5,
					SuccessThreshold:    1,
					FailureThreshold:    12,
				}, got.GetEKuiper().ReadinessProbe)
			}
		})
		t.Run("check ekuiper container liveness probe", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetEKuiper() != nil {
				got.GetEKuiper().Env = []corev1.EnvVar{}
				got.Default()
				assert.Equal(t, &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "",
							Port:   intstr.FromInt(9081),
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 10,
					TimeoutSeconds:      1,
					PeriodSeconds:       5,
					SuccessThreshold:    1,
					FailureThreshold:    12,
				}, got.GetEKuiper().LivenessProbe)
			}
		})
		t.Run("check ekuiper container probe port", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			if got.GetEKuiper() != nil {
				got.GetEKuiper().Env = []corev1.EnvVar{{Name: "KUIPER__BASIC__RESTPORT", Value: "9082"}}
				got.Default()
				assert.Equal(t, 9082, got.GetEKuiper().ReadinessProbe.HTTPGet.Port.IntValue())
				assert.Equal(t, 9082, got.GetEKuiper().ReadinessProbe.HTTPGet.Port.IntValue())
			}
		})
		t.Run("check volume template metadata", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			got.SetVolumeClaimTemplate(nil)
			got.Default()
			assert.Nil(t, got.GetVolumeClaimTemplate())

			got.SetLabels(map[string]string{"foo": "bar"})
			got.SetAnnotations(map[string]string{"foo": "bar"})
			got.SetVolumeClaimTemplate(&corev1.PersistentVolumeClaimTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"test": "fake"},
					Annotations: map[string]string{"test": "fake"},
				},
			})
			got.Default()
			assert.Equal(t, got.GetResName(), got.GetVolumeClaimTemplate().Name)
			assert.Equal(t, got.GetNamespace(), got.GetVolumeClaimTemplate().Namespace)
			assert.Equal(t, map[string]string{
				"foo":        "bar",
				"test":       "fake",
				ManagerByKey: "edge-operator",
				InstanceKey:  got.GetName(),
				ComponentKey: string(got.GetComponentType()),
			}, got.GetVolumeClaimTemplate().Labels)
			assert.Equal(t, map[string]string{
				"foo":  "bar",
				"test": "fake",
			}, got.GetVolumeClaimTemplate().Annotations)
		})
		t.Run("check service template metadata", func(t *testing.T) {
			got := deepCopyEdgeEdgeInterface(ins)
			got.SetServiceTemplate(nil)
			got.Default()
			assert.Nil(t, got.GetServiceTemplate())

			got.SetLabels(map[string]string{"foo": "bar"})
			got.SetAnnotations(map[string]string{"foo": "bar"})
			got.SetServiceTemplate(&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"test": "fake"},
					Annotations: map[string]string{"test": "fake"},
				},
			})
			got.Default()
			assert.Equal(t, got.GetResName(), got.GetServiceTemplate().Name)
			assert.Equal(t, got.GetNamespace(), got.GetServiceTemplate().Namespace)
			assert.Equal(t, map[string]string{
				"foo":        "bar",
				"test":       "fake",
				ManagerByKey: "edge-operator",
				InstanceKey:  got.GetName(),
				ComponentKey: string(got.GetComponentType()),
			}, got.GetServiceTemplate().Labels)
			assert.Equal(t, map[string]string{
				"foo":  "bar",
				"test": "fake",
			}, got.GetServiceTemplate().Annotations)
		})
		t.Run("check service template spec", func(t *testing.T) {
			if ins.GetNeuron() != nil {
				got := deepCopyEdgeEdgeInterface(ins)
				got.SetServiceTemplate(&corev1.Service{})
				got.Default()
				assert.Subset(t, got.GetServiceTemplate().Spec.Ports, []corev1.ServicePort{
					{
						Name:       "neuron",
						Protocol:   corev1.ProtocolTCP,
						Port:       7000,
						TargetPort: intstr.Parse("7000"),
					},
				})

				got.SetServiceTemplate(&corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "neuron",
								Protocol:   corev1.ProtocolTCP,
								Port:       7001,
								TargetPort: intstr.Parse("7001"),
							},
						},
					},
				})
				got.Default()
				assert.Subset(t, got.GetServiceTemplate().Spec.Ports, []corev1.ServicePort{
					{
						Name:       "neuron",
						Protocol:   corev1.ProtocolTCP,
						Port:       7001,
						TargetPort: intstr.Parse("7001"),
					},
				})
			}
			if ins.GetEKuiper() != nil {
				got := deepCopyEdgeEdgeInterface(ins)
				got.SetServiceTemplate(&corev1.Service{})
				got.Default()
				assert.Subset(t, got.GetServiceTemplate().Spec.Ports, []corev1.ServicePort{
					{
						Name:       "ekuiper",
						Protocol:   corev1.ProtocolTCP,
						Port:       9081,
						TargetPort: intstr.Parse("9081"),
					},
				})
				got.SetServiceTemplate(&corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "ekuiper",
								Protocol:   corev1.ProtocolTCP,
								Port:       9082,
								TargetPort: intstr.Parse("9082"),
							},
						},
					},
				})
				got.Default()
				assert.Subset(t, got.GetServiceTemplate().Spec.Ports, []corev1.ServicePort{
					{
						Name:       "ekuiper",
						Protocol:   corev1.ProtocolTCP,
						Port:       9082,
						TargetPort: intstr.Parse("9082"),
					},
				})
			}
		})
	}
}

func deepCopyEdgeEdgeInterface(ins EdgeInterface) EdgeInterface {
	var got EdgeInterface
	switch resource := ins.(type) {
	case *NeuronEX:
		got = resource.DeepCopy()
	case *Neuron:
		got = resource.DeepCopy()
	case *EKuiper:
		got = resource.DeepCopy()
	default:
		panic("unknown type")
	}
	return got
}
