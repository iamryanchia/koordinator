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
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	v1alpha2 "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"github.com/koordinator-sh/koordinator/apis/runtime/v1alpha1"
)

func Test_updateResource(t *testing.T) {
	type args struct {
		a *v1alpha1.LinuxContainerResources
		b *v1alpha1.LinuxContainerResources
	}
	tests := []struct {
		name string
		args args
		want *v1alpha1.LinuxContainerResources
	}{
		{
			name: "a and b are both nil",
			args: args{
				a: nil,
				b: nil,
			},
			want: nil,
		},
		{
			name: "normal case",
			args: args{
				a: &v1alpha1.LinuxContainerResources{
					CpuPeriod:              1000,
					CpuQuota:               2000,
					CpuShares:              500,
					OomScoreAdj:            10,
					MemorySwapLimitInBytes: 100,
					MemoryLimitInBytes:     300,
					CpusetCpus:             "0-64",
					CpusetMems:             "0-2",
					Unified: map[string]string{
						"resourceA": "resource A",
					},
				},
				b: &v1alpha1.LinuxContainerResources{
					CpuPeriod:              2000,
					CpuQuota:               4000,
					CpuShares:              1000,
					OomScoreAdj:            20,
					MemorySwapLimitInBytes: 200,
					MemoryLimitInBytes:     600,
					CpusetCpus:             "0-31",
					CpusetMems:             "0-4",
					Unified: map[string]string{
						"resourceB": "resource B",
					},
				},
			},
			want: &v1alpha1.LinuxContainerResources{
				CpuPeriod:              2000,
				CpuQuota:               4000,
				CpuShares:              1000,
				OomScoreAdj:            20,
				MemorySwapLimitInBytes: 200,
				MemoryLimitInBytes:     600,
				CpusetCpus:             "0-31",
				CpusetMems:             "0-4",
				Unified: map[string]string{
					"resourceA": "resource A",
					"resourceB": "resource B",
				},
			},
		},
	}
	for _, tt := range tests {
		gotResources := updateResource(tt.args.a, tt.args.b)
		assert.Equal(t, tt.want, gotResources)
	}
}

func Test_transferToKoordResources(t *testing.T) {
	type args struct {
		r *v1.LinuxContainerResources
	}
	tests := []struct {
		name string
		args args
		want *v1alpha1.LinuxContainerResources
	}{
		{
			name: "normal case",
			args: args{
				r: &v1.LinuxContainerResources{
					CpuPeriod:   1000,
					CpuShares:   500,
					OomScoreAdj: 10,
					Unified: map[string]string{
						"resourceA": "resource A",
					},
				},
			},
			want: &v1alpha1.LinuxContainerResources{
				CpuPeriod:   1000,
				CpuShares:   500,
				OomScoreAdj: 10,
				Unified: map[string]string{
					"resourceA": "resource A",
				},
			},
		},
	}
	for _, tt := range tests {
		gotResources := transferToKoordResources(tt.args.r)
		assert.Equal(t, tt.want, gotResources)
	}
}

func Test_transferToCRIResources(t *testing.T) {
	type args struct {
		r *v1alpha1.LinuxContainerResources
	}
	tests := []struct {
		name string
		args args
		want *v1.LinuxContainerResources
	}{
		{
			name: "normal case",
			args: args{
				r: &v1alpha1.LinuxContainerResources{
					CpuPeriod:   1000,
					CpuShares:   500,
					OomScoreAdj: 10,
					Unified: map[string]string{
						"resourceA": "resource A",
					},
				},
			},
			want: &v1.LinuxContainerResources{
				CpuPeriod:   1000,
				CpuShares:   500,
				OomScoreAdj: 10,
				Unified: map[string]string{
					"resourceA": "resource A",
				},
			},
		},
	}
	for _, tt := range tests {
		gotResources := transferToCRIV1alpha2Resources(tt.args.r)
		assert.Equal(t, tt.want, gotResources)
	}
}

func Test_updateResourceByUpdateContainerResourceRequest(t *testing.T) {
	type args struct {
		a *v1alpha1.LinuxContainerResources
		b *v1alpha1.LinuxContainerResources
	}
	tests := []struct {
		name string
		args args
		want *v1alpha1.LinuxContainerResources
	}{
		{
			name: "a and b are both nil",
			args: args{
				a: nil,
				b: nil,
			},
			want: nil,
		},
		{
			name: "normal case",
			args: args{
				a: &v1alpha1.LinuxContainerResources{
					CpuPeriod:              1000,
					CpuQuota:               2000,
					CpuShares:              500,
					OomScoreAdj:            10,
					MemorySwapLimitInBytes: 100,
					MemoryLimitInBytes:     300,
					CpusetCpus:             "0-64",
					CpusetMems:             "0-2",
					Unified: map[string]string{
						"resourceA": "resource A",
					},
				},
				b: &v1alpha1.LinuxContainerResources{
					CpuPeriod:              2000,
					CpuQuota:               4000,
					CpuShares:              1000,
					OomScoreAdj:            20,
					MemorySwapLimitInBytes: 200,
					MemoryLimitInBytes:     600,
					CpusetCpus:             "0-31",
					CpusetMems:             "0-4",
					Unified: map[string]string{
						"resourceB": "resource B",
					},
				},
			},
			want: &v1alpha1.LinuxContainerResources{
				CpuPeriod:              2000,
				CpuQuota:               4000,
				CpuShares:              1000,
				OomScoreAdj:            10,
				MemorySwapLimitInBytes: 200,
				MemoryLimitInBytes:     600,
				CpusetCpus:             "0-31",
				CpusetMems:             "0-4",
				Unified: map[string]string{
					"resourceA": "resource A",
					"resourceB": "resource B",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, updateResourceByUpdateContainerResourceRequest(tt.args.a, tt.args.b), "updateResourceByUpdateContainerResourceRequest(%v, %v)", tt.args.a, tt.args.b)
		})
	}
}

func Test_transferToKoordContainerEnvs(t *testing.T) {
	tests := []struct {
		name            string
		containerdEnvs  []*v1.KeyValue
		expectKoordEnvs map[string]string
	}{
		{
			name:            "containerdEnvs is nil",
			containerdEnvs:  nil,
			expectKoordEnvs: map[string]string{},
		},
		{
			name:            "containerdEnvs is not nil but with 0 item",
			containerdEnvs:  []*v1.KeyValue{},
			expectKoordEnvs: map[string]string{},
		},
		{
			name: "normal case with 1 item",
			containerdEnvs: []*v1.KeyValue{
				{
					Key:   "key1",
					Value: "value1",
				},
			},
			expectKoordEnvs: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "normal case with multi item",
			containerdEnvs: []*v1.KeyValue{
				{
					Key:   "key1",
					Value: "value1",
				},
				{
					Key:   "key2",
					Value: "value2",
				},
			},
			expectKoordEnvs: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		realKoordEnvs := transferToKoordContainerEnvs(tt.containerdEnvs)
		assert.Equalf(t, realKoordEnvs, tt.expectKoordEnvs, tt.name)
	}
}

func Test_transferToCRIContainerEnvs(t *testing.T) {
	tests := []struct {
		name                   string
		koordEnvs              map[string]string
		expectedContainerdEnvs []*v1.KeyValue
	}{
		{
			name:                   "koordEnvs is nil, should return nil",
			koordEnvs:              nil,
			expectedContainerdEnvs: nil,
		},
		{
			name:                   "koordEnvs is not nil but with 0 item, should return nil",
			koordEnvs:              map[string]string{},
			expectedContainerdEnvs: nil,
		},
		{
			name: "normal case with 1 item",
			koordEnvs: map[string]string{
				"key1": "value1",
			},
			expectedContainerdEnvs: []*v1.KeyValue{
				{
					Key:   "key1",
					Value: "value1",
				},
			},
		},
		{
			name: "normal case with multi items",
			koordEnvs: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedContainerdEnvs: []*v1.KeyValue{
				{
					Key:   "key1",
					Value: "value1",
				},
				{
					Key:   "key2",
					Value: "value2",
				},
			},
		},
	}

	for _, tt := range tests {
		realContainerdEnvs := transferToCRIV1alpha2ContainerEnvs(tt.koordEnvs)
		sort.Slice(realContainerdEnvs, func(i, j int) bool {
			return strings.Compare(realContainerdEnvs[i].GetKey(), realContainerdEnvs[j].GetKey()) > 0
		})
		sort.Slice(tt.expectedContainerdEnvs, func(i, j int) bool {
			return strings.Compare(tt.expectedContainerdEnvs[i].GetKey(), tt.expectedContainerdEnvs[j].GetKey()) > 0
		})
		assert.Equalf(t, realContainerdEnvs, tt.expectedContainerdEnvs, tt.name)
	}
}

func Test_IsV1Type(t *testing.T) {
	type args struct {
		in interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "v1 type",
			args: args{in: v1.AttachRequest{}},
			want: true,
		},
		{
			name: "v1 pointer type",
			args: args{in: &v1.AttachRequest{}},
			want: true,
		},
		{
			name: "v1alpha2 type",
			args: args{in: v1alpha2.AttachRequest{}},
			want: false,
		},
		{
			name: "v1alpha2 pointer type",
			args: args{in: &v1alpha2.AttachRequest{}},
			want: false,
		},
		{
			name: "nil",
			args: args{in: nil},
			want: false,
		},
		{
			name: "nil with v1 type",
			args: args{in: (*v1.AttachRequest)(nil)},
			want: true,
		},
		{
			name: "cutom type",
			args: args{in: struct{}{}},
			want: false,
		},
		{
			name: "cutom pointer type",
			args: args{in: &struct{}{}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsV1Type(tt.args.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_IsV1alpha2Type(t *testing.T) {
	type args struct {
		in interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "v1 type",
			args: args{in: v1.AttachRequest{}},
			want: false,
		},
		{
			name: "v1 pointer type",
			args: args{in: &v1.AttachRequest{}},
			want: false,
		},
		{
			name: "v1alpha2 type",
			args: args{in: v1alpha2.AttachRequest{}},
			want: true,
		},
		{
			name: "v1alpha2 pointer type",
			args: args{in: &v1alpha2.AttachRequest{}},
			want: true,
		},
		{
			name: "nil",
			args: args{in: nil},
			want: false,
		},
		{
			name: "nil with v1alpha2 type",
			args: args{in: (*v1alpha2.AttachRequest)(nil)},
			want: true,
		},
		{
			name: "cutom type",
			args: args{in: struct{}{}},
			want: false,
		},
		{
			name: "cutom pointer type",
			args: args{in: &struct{}{}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsV1alpha2Type(tt.args.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_V1alpha2ToV1(t *testing.T) {
	type args struct {
		in interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "v1 type",
			args:    args{in: v1.AttachRequest{}},
			want:    v1.AttachRequest{},
			wantErr: false,
		},
		{
			name:    "v1 pointer type",
			args:    args{in: &v1.AttachRequest{}},
			want:    &v1.AttachRequest{},
			wantErr: false,
		},
		{
			name:    "v1alpha2 type",
			args:    args{in: v1alpha2.AttachRequest{}},
			want:    v1alpha2.AttachRequest{},
			wantErr: true,
		},
		{
			name:    "v1alpha2 pointer type",
			args:    args{in: &v1alpha2.AttachRequest{}},
			want:    &v1.AttachRequest{},
			wantErr: false,
		},
		{
			name:    "nil",
			args:    args{in: nil},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "nil with v1 type",
			args:    args{in: (*v1.AttachRequest)(nil)},
			want:    (*v1.AttachRequest)(nil),
			wantErr: false,
		},
		{
			name:    "nil with v1alpha2 type",
			args:    args{in: (*v1alpha2.AttachRequest)(nil)},
			want:    (*v1.AttachRequest)(nil),
			wantErr: false,
		},
		{
			name:    "cutom type",
			args:    args{in: struct{}{}},
			want:    struct{}{},
			wantErr: true,
		},
		{
			name:    "cutom pointer type",
			args:    args{in: &struct{}{}},
			want:    &struct{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := V1alpha2ToV1(tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil, err)
			assert.IsType(t, tt.want, got)
		})
	}
}
