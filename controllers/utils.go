package controllers

import (
	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
)

// structAssign copy the value of struct from src to dist
func structAssign(dist, src interface{}) {
	dVal := reflect.ValueOf(dist).Elem()
	sVal := reflect.ValueOf(src).Elem()
	sType := sVal.Type()
	for i := 0; i < sVal.NumField(); i++ {
		// we need to check if the dist struct has the same field
		name := sType.Field(i).Name
		if ok := dVal.FieldByName(name).IsValid(); ok {
			dVal.FieldByName(name).Set(reflect.ValueOf(sVal.Field(i).Interface()))
		}
	}
}

// usePVC determines whether we should attach a PVC to a pod.
func usePVC(ins edgev1alpha1.EdgeInterface) bool {
	var storage *resource.Quantity

	claim := ins.GetVolumeClaimTemplate()
	if claim != nil {
		requests := claim.Spec.Resources.Requests
		if requests != nil {
			storageCopy := requests[corev1.ResourceStorage]
			storage = &storageCopy
		}
	}
	return storage != nil && !storage.IsZero()
}
