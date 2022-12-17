package controllers

import (
	"reflect"

	edgev1alpha1 "github.com/emqx/edge-operator/api/v1alpha1"
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

// mergeVolumes merge the volumes of enuron and ekuiper container
func mergeVolumes(vols *[]volumeInfo) {
	ts := []edgev1alpha1.ComponentType{edgev1alpha1.ComponentTypeNeuron, edgev1alpha1.ComponentTypeEKuiper}
	for _, t := range ts {
		*vols = append(*vols, defaultVolume[t]...)
	}
}
