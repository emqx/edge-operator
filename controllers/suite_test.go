/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

var ctx context.Context
var cancel context.CancelFunc
var timeout, interval time.Duration

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())
	timeout = time.Second * 15
	interval = time.Millisecond * 250

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = edgev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(NewNeuronEXReconciler(mgr).SetupWithManager(mgr)).Should(Succeed())
	Expect(NewNeuronReconciler(mgr).SetupWithManager(mgr)).Should(Succeed())
	Expect(NewEKuiperReconciler(mgr).SetupWithManager(mgr)).Should(Succeed())

	go func() {
		defer GinkgoRecover()
		//https://github.com/hazelcast/hazelcast-platform-operator/commit/1fa58002a4f567ef4d6c2f53c919ad6f84d8bbc1
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	// https://github.com/kubernetes-sigs/controller-runtime/issues/1571
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func getNeuronEX() *edgev1alpha1.NeuronEX {
	neuronEX := &edgev1alpha1.NeuronEX{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "neuronex",
			Namespace: "default",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"foo": "bar",
			},
		},
		Spec: edgev1alpha1.NeuronEXSpec{
			Neuron: corev1.Container{
				Name:  "neuron",
				Image: "emqx/neuron:2.3",
			},
			EKuiper: corev1.Container{
				Name:  "ekuiper",
				Image: "lfedge/ekuiper:1.8-slim",
			},
		},
	}
	neuronEX.Default()
	return neuronEX
}

func getNeuron() *edgev1alpha1.Neuron {
	neuron := &edgev1alpha1.Neuron{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "neuron",
			Namespace: "default",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"foo": "bar",
			},
		},
		Spec: edgev1alpha1.NeuronSpec{
			Neuron: corev1.Container{
				Name:  "neuron",
				Image: "emqx/neuron:2.3",
			},
		},
	}
	neuron.Default()
	return neuron
}

func getEKuiper() *edgev1alpha1.EKuiper {
	ekuiper := &edgev1alpha1.EKuiper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ekuiper",
			Namespace: "default",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"foo": "bar",
			},
		},
		Spec: edgev1alpha1.EKuiperSpec{
			EKuiper: corev1.Container{
				Name:  "ekuiper",
				Image: "lfedge/ekuiper:1.8-slim",
			},
		},
	}
	ekuiper.Default()
	return ekuiper
}

func deepCopyEdgeEdgeInterface(ins edgev1alpha1.EdgeInterface) edgev1alpha1.EdgeInterface {
	var got edgev1alpha1.EdgeInterface
	switch resource := ins.(type) {
	case *edgev1alpha1.NeuronEX:
		got = resource.DeepCopy()
	case *edgev1alpha1.Neuron:
		got = resource.DeepCopy()
	case *edgev1alpha1.EKuiper:
		got = resource.DeepCopy()
	default:
		panic("unknown type")
	}
	return got
}

func addVolumeTemplate(ins edgev1alpha1.EdgeInterface) edgev1alpha1.EdgeInterface {
	volumeTemplate := &corev1.PersistentVolumeClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"e2e/test": "volumeTemplate",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("8Mi"),
				},
			},
		},
	}
	switch resource := ins.(type) {
	case *edgev1alpha1.NeuronEX:
		new := resource.DeepCopy()
		new.Spec.VolumeClaimTemplate = volumeTemplate
		new.Default()
		return new
	case *edgev1alpha1.Neuron:
		new := resource.DeepCopy()
		new.Spec.VolumeClaimTemplate = volumeTemplate
		new.Default()
		return new
	case *edgev1alpha1.EKuiper:
		new := resource.DeepCopy()
		new.Spec.VolumeClaimTemplate = volumeTemplate
		new.Default()
		return new
	default:
		panic("unknown type")
	}
}

func addServiceTemplate(ins edgev1alpha1.EdgeInterface) edgev1alpha1.EdgeInterface {
	serviceTemplate := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"e2e/test": "serviceTemplate",
			},
		},
	}
	switch resource := ins.(type) {
	case *edgev1alpha1.NeuronEX:
		new := resource.DeepCopy()
		new.Spec.ServiceTemplate = serviceTemplate
		new.Default()
		return new
	case *edgev1alpha1.Neuron:
		new := resource.DeepCopy()
		new.Spec.ServiceTemplate = serviceTemplate
		new.Default()
		return new
	case *edgev1alpha1.EKuiper:
		new := resource.DeepCopy()
		new.Spec.ServiceTemplate = serviceTemplate
		new.Default()
		return new
	default:
		panic("unknown type")
	}
}
