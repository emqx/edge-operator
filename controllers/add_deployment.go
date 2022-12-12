package controllers

import (
	"context"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	"github.com/emqx/edge-operator/internal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type addEkuiperDeployment struct{}

func (a addEkuiperDeployment) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.EKuiper) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add ekuiper Deployment")

	deploy := getDeployment(instance)
	if err := r.createOrUpdate(ctx, instance, &deploy, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}

type addNeuronDeployment struct{}

func (a addNeuronDeployment) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.Neuron) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler", "add Neuron Deployment")

	deploy := getDeployment(instance)
	if err := r.createOrUpdate(ctx, instance, &deploy, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}

type addNeuronExDeploy struct{}

func (a addNeuronExDeploy) reconcile(ctx context.Context, r *EdgeController, instance *edgev1alpha1.NeuronEX) *requeue {
	logger := log.WithValues("namespace", instance.Namespace, "instance", instance.Name, "reconciler",
		"add NeuronEx Deploy")

	deploy := getDeployment(instance)
	if err := r.createOrUpdate(ctx, instance, &deploy, logger); err != nil {
		return &requeue{curError: err}
	}
	return nil
}

func getDeployment(instance edgev1alpha1.EdgeInterface) appsv1.Deployment {
	podTemp := getPodTemplate(instance)

	deploy := internal.GetDeployment(instance, instance.GetComponentType(), &podTemp)
	return deploy
}

func getPodTemplate(instance edgev1alpha1.EdgeInterface) corev1.PodTemplateSpec {
	compType := instance.GetComponentType()

	pod := corev1.PodTemplateSpec{
		ObjectMeta: internal.GetObjectMetadata(instance, nil, compType),
		Spec:       getPodSpec(instance),
	}

	// TODO: return spec only after set default label by webhook
	pod.Spec.Volumes = append(pod.Spec.Volumes, getVolumes(instance)...)
	return pod
}

func getPodSpec(instance edgev1alpha1.EdgeInterface) corev1.PodSpec {
	podSpec := &corev1.PodSpec{}
	edgePodSpec := instance.GetEdgePodSpec()
	structAssign(podSpec, &edgePodSpec)

	switch instance.GetComponentType() {
	case edgev1alpha1.ComponentTypeNeuronEx:
		podSpec.Containers = []corev1.Container{
			getNeuronContainer(instance, instance.GetNeuron()),
			getEkuiperContainer(instance, instance.GetEKuiper()),
			getEkuiperToolContainer(instance.GetEKuiper()),
		}
	case edgev1alpha1.ComponentTypeEKuiper:
		podSpec.Containers = []corev1.Container{
			getEkuiperContainer(instance, instance.GetEKuiper()),
		}
	case edgev1alpha1.ComponentTypeNeuron:
		podSpec.Containers = []corev1.Container{
			getNeuronContainer(instance, instance.GetNeuron()),
		}
	default:
		panic("unknown component " + instance.GetComponentType())
	}
	return *podSpec
}

func getVolumes(ins edgev1alpha1.EdgeInterface) (volumes []corev1.Volume) {
	volumes = make([]corev1.Volume, 0)

	compType := ins.GetComponentType()
	pvcs := defaultPVC[compType]
	for _, pvc := range pvcs {
		volume := corev1.Volume{
			Name: pvc.name,
		}
		if usePVC(ins) {
			volume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: internal.GetPvcName(ins, pvc.name),
			}
		} else {
			volume.VolumeSource = corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}
		}
		volumes = append(volumes, volume)
	}

	if ins.GetComponentType() == edgev1alpha1.ComponentTypeNeuronEx {
		cm := internal.ConfigMaps[internal.EKuiperToolConfig]
		volumes = append(volumes, internal.GetVolume(ins, cm))
	}
	return
}

func getNeuronContainer(ins edgev1alpha1.EdgeInterface, conf *corev1.Container) corev1.Container {
	container := conf.DeepCopy()
	if container.Name == "" {
		container.Name = ins.GetComponentType().String()
	}

	var pvcs []pvcInfo
	if ins.GetComponentType() == edgev1alpha1.ComponentTypeNeuronEx {
		pvcs = defaultPVC[edgev1alpha1.ComponentTypeNeuronEx]
	} else {
		pvcs = defaultPVC[edgev1alpha1.ComponentTypeNeuron]
	}
	for i := range pvcs {
		container.VolumeMounts = append(container.VolumeMounts,
			corev1.VolumeMount{
				Name:      pvcs[i].name,
				MountPath: pvcs[i].mountPath,
			})
	}
	return *container
}

func getEkuiperContainer(ins edgev1alpha1.EdgeInterface, conf *corev1.Container) corev1.Container {
	container := conf.DeepCopy()
	if container.Name == "" {
		container.Name = ins.GetComponentType().String()
	}

	// TODO: add default value in webhook
	extendEnv(container, []corev1.EnvVar{
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
	})

	pvcs := defaultPVC[edgev1alpha1.ComponentTypeEKuiper]
	for i := range pvcs {
		container.VolumeMounts = append(container.VolumeMounts,
			corev1.VolumeMount{
				Name:      pvcs[i].name,
				MountPath: pvcs[i].mountPath,
			})
	}
	return *container
}

func getEkuiperToolContainer(conf *corev1.Container) corev1.Container {
	cmi := internal.ConfigMaps[internal.EKuiperToolConfig]

	// TODO: Is it the latest version of eKuiper tool compatible with the eKuiper that user specifies?
	container := corev1.Container{
		Name:            "ekuiper-tool",
		Image:           "lfedge/ekuiper-kubernetes-tool:latest",
		ImagePullPolicy: conf.ImagePullPolicy,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      cmi.MountName,
				MountPath: cmi.MountPath,
				ReadOnly:  true,
			},
		},
	}
	return container
}
