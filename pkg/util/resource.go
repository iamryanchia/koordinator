/*
Copyright 2022 The Koordinator Authors.

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

package util

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	quotav1 "k8s.io/apiserver/pkg/quota/v1"
)

func NewZeroResourceList() corev1.ResourceList {
	return corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(0, resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(0, resource.BinarySI),
	}
}

// MultiplyMilliQuant scales quantity by factor
func MultiplyMilliQuant(quant resource.Quantity, factor float64) resource.Quantity {
	milliValue := quant.MilliValue()
	newMilliValue := int64(float64(milliValue) * factor)
	newQuant := resource.NewMilliQuantity(newMilliValue, quant.Format)
	return *newQuant
}

// MultiplyQuant scales quantity by factor
func MultiplyQuant(quant resource.Quantity, factor float64) resource.Quantity {
	value := quant.Value()
	newValue := int64(float64(value) * factor)
	newQuant := resource.NewQuantity(newValue, quant.Format)
	return *newQuant
}

// MinResourceList returns the result of Min(a, b) for each named resource, corresponding to the quotav1.Max().
// It should be semantically equivalent to the result of `quotav1.Subtract(quotav1.Add(a, b), quotav1.Max(a, b))`.
//
// e.g.
//
//	a = {"cpu": "10", "memory": "20Gi"}, b = {"cpu": "6", "memory": "24Gi", "nvidia.com/gpu": "2"}
//	=> {"cpu": "6", "memory": "20Gi"}
func MinResourceList(a corev1.ResourceList, b corev1.ResourceList) corev1.ResourceList {
	result := corev1.ResourceList{}
	for key, value := range a {
		other, found := b[key]
		if !found {
			continue
		}
		if value.Cmp(other) >= 0 {
			result[key] = other.DeepCopy()
		} else {
			result[key] = value.DeepCopy()
		}
	}
	return result
}

// IsResourceListEqualValue checks if the two resource lists are numerically equivalent.
// NOTE: Resource name with a zero value will be ignored in comparison.
// e.g. a = {"cpu": "10", "memory": "0"}, b = {"cpu": "10"} => true
func IsResourceListEqualValue(a, b corev1.ResourceList) bool {
	a = quotav1.RemoveZeros(a)
	b = quotav1.RemoveZeros(b)
	if len(a) != len(b) { // different number of no-zero resources
		return false
	}
	for key, value := range a {
		other, found := b[key]
		if !found { // different resource names since all zero values have been dropped
			return false
		}
		if value.Cmp(other) != 0 { // different values
			return false
		}
	}
	return true
}

// IsResourceDiff returns whether the new resource has big enough difference with the old one or not
func IsResourceDiff(old, new corev1.ResourceList, resourceName corev1.ResourceName, diffThreshold float64) bool {
	oldQuant := 0.0
	oldResource, oldExist := old[resourceName]
	if oldExist {
		oldQuant = float64(oldResource.MilliValue())
	}

	newQuant := 0.0
	newResource, newExist := new[resourceName]
	if newExist {
		newQuant = float64(newResource.MilliValue())
	}

	if oldExist != newExist {
		return true
	}

	if !oldExist && !newExist {
		return false
	}

	// not equal for both are zero
	return newQuant > oldQuant*(1+diffThreshold) || newQuant < oldQuant*(1-diffThreshold)
}

func QuantityPtr(q resource.Quantity) *resource.Quantity {
	return &q
}
