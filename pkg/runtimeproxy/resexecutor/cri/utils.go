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

package cri

import (
	"errors"
	"fmt"
	"reflect"

	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	v1alpha2 "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"github.com/koordinator-sh/koordinator/apis/runtime/v1alpha1"
	"github.com/koordinator-sh/koordinator/pkg/runtimeproxy/utils"
)

func transferToKoordResources(r *v1.LinuxContainerResources) *v1alpha1.LinuxContainerResources {
	linuxResources := &v1alpha1.LinuxContainerResources{
		CpuPeriod:              r.GetCpuPeriod(),
		CpuQuota:               r.GetCpuQuota(),
		CpuShares:              r.GetCpuShares(),
		MemoryLimitInBytes:     r.GetMemoryLimitInBytes(),
		OomScoreAdj:            r.GetOomScoreAdj(),
		CpusetCpus:             r.GetCpusetCpus(),
		CpusetMems:             r.GetCpusetMems(),
		Unified:                r.GetUnified(),
		MemorySwapLimitInBytes: r.GetMemorySwapLimitInBytes(),
	}

	for _, item := range r.GetHugepageLimits() {
		linuxResources.HugepageLimits = append(linuxResources.HugepageLimits, &v1alpha1.HugepageLimit{
			PageSize: item.GetPageSize(),
			Limit:    item.GetLimit(),
		})
	}
	return linuxResources
}

func transferToCRIV1alpha2Resources(r *v1alpha1.LinuxContainerResources) *v1alpha2.LinuxContainerResources {
	linuxResources := &v1alpha2.LinuxContainerResources{
		CpuPeriod:              r.GetCpuPeriod(),
		CpuQuota:               r.GetCpuQuota(),
		CpuShares:              r.GetCpuShares(),
		MemoryLimitInBytes:     r.GetMemoryLimitInBytes(),
		OomScoreAdj:            r.GetOomScoreAdj(),
		CpusetCpus:             r.GetCpusetCpus(),
		CpusetMems:             r.GetCpusetMems(),
		Unified:                r.GetUnified(),
		MemorySwapLimitInBytes: r.GetMemorySwapLimitInBytes(),
	}

	for _, item := range r.GetHugepageLimits() {
		linuxResources.HugepageLimits = append(linuxResources.HugepageLimits, &v1alpha2.HugepageLimit{
			PageSize: item.GetPageSize(),
			Limit:    item.GetLimit(),
		})
	}
	return linuxResources
}

func updateResource(a, b *v1alpha1.LinuxContainerResources) *v1alpha1.LinuxContainerResources {
	if a == nil || b == nil {
		return a
	}
	if b.CpuPeriod > 0 {
		a.CpuPeriod = b.CpuPeriod
	}
	if b.CpuQuota != 0 { // -1 is valid
		a.CpuQuota = b.CpuQuota
	}
	if b.CpuShares > 0 {
		a.CpuShares = b.CpuShares
	}
	if b.MemoryLimitInBytes > 0 {
		a.MemoryLimitInBytes = b.MemoryLimitInBytes
	}
	if b.OomScoreAdj >= -1000 && b.OomScoreAdj <= 1000 {
		a.OomScoreAdj = b.OomScoreAdj
	}

	a.CpusetCpus = b.CpusetCpus
	a.CpusetMems = b.CpusetMems

	a.Unified = utils.MergeMap(a.Unified, b.Unified)
	if b.MemorySwapLimitInBytes > 0 {
		a.MemorySwapLimitInBytes = b.MemorySwapLimitInBytes
	}
	return a
}

// updateResourceByUpdateContainerResourceRequest updates resources in cache by UpdateContainerResource request.
// updateResourceByUpdateContainerResourceRequest will omit OomScoreAdj.
//
// Normally kubelet won't send UpdateContainerResource request, so if some components want to send it and want to update OomScoreAdj,
// please use hook to achieve it.
func updateResourceByUpdateContainerResourceRequest(a, b *v1alpha1.LinuxContainerResources) *v1alpha1.LinuxContainerResources {
	if a == nil || b == nil {
		return a
	}
	if b.CpuPeriod > 0 {
		a.CpuPeriod = b.CpuPeriod
	}
	if b.CpuQuota != 0 { // -1 is valid
		a.CpuQuota = b.CpuQuota
	}
	if b.CpuShares > 0 {
		a.CpuShares = b.CpuShares
	}
	if b.MemoryLimitInBytes > 0 {
		a.MemoryLimitInBytes = b.MemoryLimitInBytes
	}
	if b.CpusetCpus != "" {
		a.CpusetCpus = b.CpusetCpus
	}
	if b.CpusetMems != "" {
		a.CpusetMems = b.CpusetMems
	}
	a.Unified = utils.MergeMap(a.Unified, b.Unified)
	if b.MemorySwapLimitInBytes > 0 {
		a.MemorySwapLimitInBytes = b.MemorySwapLimitInBytes
	}
	return a
}

func transferToKoordContainerEnvs(envs []*v1.KeyValue) map[string]string {
	res := make(map[string]string)
	if envs == nil {
		return res
	}
	for _, item := range envs {
		res[item.GetKey()] = item.GetValue()
	}
	return res
}

func transferToCRIV1alpha2ContainerEnvs(envs map[string]string) []*v1alpha2.KeyValue {
	var res []*v1alpha2.KeyValue
	if envs == nil {
		return res
	}
	for key, val := range envs {
		res = append(res, &v1alpha2.KeyValue{
			Key:   key,
			Value: val,
		})
	}
	return res
}

func transferToCRIV1ContainerEnvs(envs map[string]string) []*v1.KeyValue {
	var res []*v1.KeyValue
	if envs == nil {
		return res
	}
	for key, val := range envs {
		res = append(res, &v1.KeyValue{
			Key:   key,
			Value: val,
		})
	}
	return res
}

func IsKeyValExistInLabels(labels map[string]string, key, val string) bool {
	if labels == nil {
		return false
	}
	for curKey, curVal := range labels {
		if curKey == key && curVal == val {
			return true
		}
	}
	return false
}

type Marshaler interface {
	Marshal() ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal([]byte) error
}

func ConvertByMarshal(src Marshaler, dst Unmarshaler) error {
	data, err := src.Marshal()
	if err != nil {
		return err
	}

	return dst.Unmarshal(data)
}

var (
	v1PkgPath       = reflect.TypeOf(v1.AttachRequest{}).PkgPath()
	v1alpha2PkgPath = reflect.TypeOf(v1alpha2.AttachRequest{}).PkgPath()

	typeMapper = map[reflect.Type]reflect.Type{
		reflect.TypeOf(v1alpha2.AttachRequest{}):                    reflect.TypeOf(v1.AttachRequest{}),
		reflect.TypeOf(v1alpha2.AttachRequest{}):                    reflect.TypeOf(v1.AttachRequest{}),
		reflect.TypeOf(v1alpha2.AttachResponse{}):                   reflect.TypeOf(v1.AttachResponse{}),
		reflect.TypeOf(v1alpha2.ContainerStatsRequest{}):            reflect.TypeOf(v1.ContainerStatsRequest{}),
		reflect.TypeOf(v1alpha2.ContainerStatsResponse{}):           reflect.TypeOf(v1.ContainerStatsResponse{}),
		reflect.TypeOf(v1alpha2.ContainerStatusRequest{}):           reflect.TypeOf(v1.ContainerStatusRequest{}),
		reflect.TypeOf(v1alpha2.ContainerStatusResponse{}):          reflect.TypeOf(v1.ContainerStatusResponse{}),
		reflect.TypeOf(v1alpha2.CreateContainerRequest{}):           reflect.TypeOf(v1.CreateContainerRequest{}),
		reflect.TypeOf(v1alpha2.CreateContainerResponse{}):          reflect.TypeOf(v1.CreateContainerResponse{}),
		reflect.TypeOf(v1alpha2.ExecRequest{}):                      reflect.TypeOf(v1.ExecRequest{}),
		reflect.TypeOf(v1alpha2.ExecResponse{}):                     reflect.TypeOf(v1.ExecResponse{}),
		reflect.TypeOf(v1alpha2.ExecSyncRequest{}):                  reflect.TypeOf(v1.ExecSyncRequest{}),
		reflect.TypeOf(v1alpha2.ExecSyncResponse{}):                 reflect.TypeOf(v1.ExecSyncResponse{}),
		reflect.TypeOf(v1alpha2.ImageFsInfoRequest{}):               reflect.TypeOf(v1.ImageFsInfoRequest{}),
		reflect.TypeOf(v1alpha2.ImageFsInfoResponse{}):              reflect.TypeOf(v1.ImageFsInfoResponse{}),
		reflect.TypeOf(v1alpha2.ImageStatusRequest{}):               reflect.TypeOf(v1.ImageStatusRequest{}),
		reflect.TypeOf(v1alpha2.ImageStatusResponse{}):              reflect.TypeOf(v1.ImageStatusResponse{}),
		reflect.TypeOf(v1alpha2.ListContainersRequest{}):            reflect.TypeOf(v1.ListContainersRequest{}),
		reflect.TypeOf(v1alpha2.ListContainersResponse{}):           reflect.TypeOf(v1.ListContainersResponse{}),
		reflect.TypeOf(v1alpha2.ListContainerStatsRequest{}):        reflect.TypeOf(v1.ListContainerStatsRequest{}),
		reflect.TypeOf(v1alpha2.ListContainerStatsResponse{}):       reflect.TypeOf(v1.ListContainerStatsResponse{}),
		reflect.TypeOf(v1alpha2.ListImagesRequest{}):                reflect.TypeOf(v1.ListImagesRequest{}),
		reflect.TypeOf(v1alpha2.ListImagesResponse{}):               reflect.TypeOf(v1.ListImagesResponse{}),
		reflect.TypeOf(v1alpha2.ListPodSandboxRequest{}):            reflect.TypeOf(v1.ListPodSandboxRequest{}),
		reflect.TypeOf(v1alpha2.ListPodSandboxResponse{}):           reflect.TypeOf(v1.ListPodSandboxResponse{}),
		reflect.TypeOf(v1alpha2.ListPodSandboxStatsRequest{}):       reflect.TypeOf(v1.ListPodSandboxStatsRequest{}),
		reflect.TypeOf(v1alpha2.ListPodSandboxStatsResponse{}):      reflect.TypeOf(v1.ListPodSandboxStatsResponse{}),
		reflect.TypeOf(v1alpha2.PodSandboxStatsRequest{}):           reflect.TypeOf(v1.PodSandboxStatsRequest{}),
		reflect.TypeOf(v1alpha2.PodSandboxStatsResponse{}):          reflect.TypeOf(v1.PodSandboxStatsResponse{}),
		reflect.TypeOf(v1alpha2.PodSandboxStatusRequest{}):          reflect.TypeOf(v1.PodSandboxStatusRequest{}),
		reflect.TypeOf(v1alpha2.PodSandboxStatusResponse{}):         reflect.TypeOf(v1.PodSandboxStatusResponse{}),
		reflect.TypeOf(v1alpha2.PortForwardRequest{}):               reflect.TypeOf(v1.PortForwardRequest{}),
		reflect.TypeOf(v1alpha2.PortForwardResponse{}):              reflect.TypeOf(v1.PortForwardResponse{}),
		reflect.TypeOf(v1alpha2.PullImageRequest{}):                 reflect.TypeOf(v1.PullImageRequest{}),
		reflect.TypeOf(v1alpha2.PullImageResponse{}):                reflect.TypeOf(v1.PullImageResponse{}),
		reflect.TypeOf(v1alpha2.RemoveContainerRequest{}):           reflect.TypeOf(v1.RemoveContainerRequest{}),
		reflect.TypeOf(v1alpha2.RemoveContainerResponse{}):          reflect.TypeOf(v1.RemoveContainerResponse{}),
		reflect.TypeOf(v1alpha2.RemoveImageRequest{}):               reflect.TypeOf(v1.RemoveImageRequest{}),
		reflect.TypeOf(v1alpha2.RemoveImageResponse{}):              reflect.TypeOf(v1.RemoveImageResponse{}),
		reflect.TypeOf(v1alpha2.RemovePodSandboxRequest{}):          reflect.TypeOf(v1.RemovePodSandboxRequest{}),
		reflect.TypeOf(v1alpha2.RemovePodSandboxResponse{}):         reflect.TypeOf(v1.RemovePodSandboxResponse{}),
		reflect.TypeOf(v1alpha2.ReopenContainerLogRequest{}):        reflect.TypeOf(v1.ReopenContainerLogRequest{}),
		reflect.TypeOf(v1alpha2.ReopenContainerLogResponse{}):       reflect.TypeOf(v1.ReopenContainerLogResponse{}),
		reflect.TypeOf(v1alpha2.RunPodSandboxRequest{}):             reflect.TypeOf(v1.RunPodSandboxRequest{}),
		reflect.TypeOf(v1alpha2.RunPodSandboxResponse{}):            reflect.TypeOf(v1.RunPodSandboxResponse{}),
		reflect.TypeOf(v1alpha2.StartContainerRequest{}):            reflect.TypeOf(v1.StartContainerRequest{}),
		reflect.TypeOf(v1alpha2.StartContainerResponse{}):           reflect.TypeOf(v1.StartContainerResponse{}),
		reflect.TypeOf(v1alpha2.StatusRequest{}):                    reflect.TypeOf(v1.StatusRequest{}),
		reflect.TypeOf(v1alpha2.StatusResponse{}):                   reflect.TypeOf(v1.StatusResponse{}),
		reflect.TypeOf(v1alpha2.StopContainerRequest{}):             reflect.TypeOf(v1.StopContainerRequest{}),
		reflect.TypeOf(v1alpha2.StopContainerResponse{}):            reflect.TypeOf(v1.StopContainerResponse{}),
		reflect.TypeOf(v1alpha2.StopPodSandboxRequest{}):            reflect.TypeOf(v1.StopPodSandboxRequest{}),
		reflect.TypeOf(v1alpha2.StopPodSandboxResponse{}):           reflect.TypeOf(v1.StopPodSandboxResponse{}),
		reflect.TypeOf(v1alpha2.UpdateContainerResourcesRequest{}):  reflect.TypeOf(v1.UpdateContainerResourcesRequest{}),
		reflect.TypeOf(v1alpha2.UpdateContainerResourcesResponse{}): reflect.TypeOf(v1.UpdateContainerResourcesResponse{}),
		reflect.TypeOf(v1alpha2.UpdateRuntimeConfigRequest{}):       reflect.TypeOf(v1.UpdateRuntimeConfigRequest{}),
		reflect.TypeOf(v1alpha2.UpdateRuntimeConfigResponse{}):      reflect.TypeOf(v1.UpdateRuntimeConfigResponse{}),
		reflect.TypeOf(v1alpha2.VersionRequest{}):                   reflect.TypeOf(v1.VersionRequest{}),
		reflect.TypeOf(v1alpha2.VersionResponse{}):                  reflect.TypeOf(v1.VersionResponse{}),
		// The following structs are nested by the above requests or responses.
		reflect.TypeOf(v1alpha2.PodSandbox{}):              reflect.TypeOf(v1.PodSandbox{}),
		reflect.TypeOf(v1alpha2.Container{}):               reflect.TypeOf(v1.Container{}),
		reflect.TypeOf(v1alpha2.LinuxContainerResources{}): reflect.TypeOf(v1.LinuxContainerResources{}),
		reflect.TypeOf(v1alpha2.KeyValue{}):                reflect.TypeOf(v1.KeyValue{}),
	}
)

// V1alpha2ToV1 converts a v1alpha2 object pointer to corresponding v1 object pointer.
// If the passed value belongs to the v1 API, the value will be returned as is.
func V1alpha2ToV1(in interface{}) (interface{}, error) {
	if in == nil {
		return nil, errors.New("nil pointer")
	}

	if IsV1Type(in) {
		return in, nil
	}

	inType := reflect.TypeOf(in)
	if inType.Kind() != reflect.Ptr {
		// Only pointer type has a marshal method.
		return in, fmt.Errorf("not a pointer type")
	}
	inType = inType.Elem()

	if inType.PkgPath() != v1alpha2PkgPath {
		return in, fmt.Errorf("not CRI v1alpha2 API object: %v", reflect.TypeOf(in))
	}

	outType, ok := typeMapper[inType]
	if !ok {
		return in, fmt.Errorf("undefined type conversion: %v", reflect.TypeOf(in))
	}

	outPtr := reflect.New(outType)
	if !reflect.ValueOf(in).IsNil() {
		if err := ConvertByMarshal(in.(Marshaler), outPtr.Interface().(Unmarshaler)); err != nil {
			return in, err
		}
	}

	return outPtr.Interface(), nil
}

// IsV1Type checks whether the passed object is the type of CRI v1 API.
func IsV1Type(in interface{}) bool {
	if in == nil {
		return false
	}

	inType := reflect.TypeOf(in)
	if inType.Kind() == reflect.Ptr {
		inType = inType.Elem()
	}

	return inType.PkgPath() == v1PkgPath
}

// IsV1alpha2Type checks whether the passed object is the type of CRI v1alpha2 API.
func IsV1alpha2Type(in interface{}) bool {
	if in == nil {
		return false
	}

	inType := reflect.TypeOf(in)
	if inType.Kind() == reflect.Ptr {
		inType = inType.Elem()
	}

	return inType.PkgPath() == v1alpha2PkgPath
}
