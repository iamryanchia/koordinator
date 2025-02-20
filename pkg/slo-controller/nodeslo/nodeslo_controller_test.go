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

package nodeslo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/koordinator-sh/koordinator/apis/configuration"
	slov1alpha1 "github.com/koordinator-sh/koordinator/apis/slo/v1alpha1"
	"github.com/koordinator-sh/koordinator/pkg/util/sloconfig"
)

// change strategy to interface
// plugins in koordlet/resmanager will get the interface and
// change it to strategy individually
func getExtensionsIfMap(in slov1alpha1.ExtensionsMap) (*slov1alpha1.ExtensionsMap, error) {
	extensionsMap := &slov1alpha1.ExtensionsMap{Object: map[string]interface{}{}}
	for extkey, extIf := range in.Object {
		//marshal unmarshal to
		extStr, err := json.Marshal(extIf)
		if err != nil {
			return nil, err
		}
		var strategy interface{}
		if err := json.Unmarshal(extStr, &strategy); err != nil {
			return nil, err
		}
		if extensionsMap.Object == nil {
			extensionsMap.Object = make(map[string]interface{})
		}
		extensionsMap.Object[extkey] = strategy
	}

	return extensionsMap, nil
}

func TestNodeSLOReconciler_initNodeSLO(t *testing.T) {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	slov1alpha1.AddToScheme(scheme)
	testingResourceThresholdStrategy := sloconfig.DefaultResourceThresholdStrategy()
	testingResourceThresholdStrategy.CPUSuppressThresholdPercent = pointer.Int64(60)
	testingResourceQOSStrategyOld := &slov1alpha1.ResourceQOSStrategy{
		BEClass: &slov1alpha1.ResourceQOS{
			CPUQOS: &slov1alpha1.CPUQOSCfg{
				CPUQOS: slov1alpha1.CPUQOS{
					GroupIdentity: pointer.Int64(0),
				},
			},
		},
	}
	testingResourceQOSStrategy := &slov1alpha1.ResourceQOSStrategy{
		BEClass: &slov1alpha1.ResourceQOS{
			CPUQOS: &slov1alpha1.CPUQOSCfg{
				CPUQOS: slov1alpha1.CPUQOS{
					GroupIdentity: pointer.Int64(0),
				},
			},
		},
	}
	testingExtensions := getDefaultExtensionStrategy()
	type args struct {
		node    *corev1.Node
		nodeSLO *slov1alpha1.NodeSLO
	}
	type fields struct {
		configMap *corev1.ConfigMap
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *slov1alpha1.NodeSLOSpec
		wantErr bool
	}{
		{
			name: "throw an error if no slo configmap",
			args: args{
				node:    &corev1.Node{},
				nodeSLO: &slov1alpha1.NodeSLO{},
			},
			fields: fields{},
			want: &slov1alpha1.NodeSLOSpec{
				ResourceUsedThresholdWithBE: sloconfig.DefaultResourceThresholdStrategy(),
				ResourceQOSStrategy:         &slov1alpha1.ResourceQOSStrategy{},
				CPUBurstStrategy:            sloconfig.DefaultCPUBurstStrategy(),
				SystemStrategy:              sloconfig.DefaultSystemStrategy(),
				Extensions:                  testingExtensions,
			},
			wantErr: false,
		},
		{
			name: "unmarshal failed, use the default",
			args: args{
				node:    &corev1.Node{},
				nodeSLO: &slov1alpha1.NodeSLO{},
			},
			fields: fields{configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sloconfig.SLOCtrlConfigMap,
					Namespace: sloconfig.ConfigNameSpace,
				},
				Data: map[string]string{
					configuration.ResourceThresholdConfigKey: "{\"clusterStrategy\":{\"invalidField\",\"cpuSuppressThresholdPercent\":60}}",
					configuration.ResourceQOSConfigKey:       "{\"clusterStrategy\":{\"invalidField\"}}",
				},
			}},
			want: &slov1alpha1.NodeSLOSpec{
				ResourceUsedThresholdWithBE: sloconfig.DefaultResourceThresholdStrategy(),
				ResourceQOSStrategy:         &slov1alpha1.ResourceQOSStrategy{},
				CPUBurstStrategy:            sloconfig.DefaultCPUBurstStrategy(),
				SystemStrategy:              sloconfig.DefaultSystemStrategy(),
				Extensions:                  testingExtensions,
			},
			wantErr: false,
		},
		{
			name: "get spec successfully",
			args: args{
				node:    &corev1.Node{},
				nodeSLO: &slov1alpha1.NodeSLO{},
			},
			fields: fields{configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sloconfig.SLOCtrlConfigMap,
					Namespace: sloconfig.ConfigNameSpace,
				},
				Data: map[string]string{
					configuration.ResourceThresholdConfigKey: "{\"clusterStrategy\":{\"enable\":false,\"cpuSuppressThresholdPercent\":60}}",
				},
			}},
			want: &slov1alpha1.NodeSLOSpec{
				ResourceUsedThresholdWithBE: testingResourceThresholdStrategy,
				ResourceQOSStrategy:         &slov1alpha1.ResourceQOSStrategy{},
				CPUBurstStrategy:            sloconfig.DefaultCPUBurstStrategy(),
				SystemStrategy:              sloconfig.DefaultSystemStrategy(),
				Extensions:                  testingExtensions,
			},
			wantErr: false,
		},
		{
			name: "get spec successfully 1",
			args: args{
				node:    &corev1.Node{},
				nodeSLO: &slov1alpha1.NodeSLO{},
			},
			fields: fields{configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sloconfig.SLOCtrlConfigMap,
					Namespace: sloconfig.ConfigNameSpace,
				},
				Data: map[string]string{
					configuration.ResourceThresholdConfigKey: "{\"clusterStrategy\":{\"enable\":false,\"cpuSuppressThresholdPercent\":60}}",
					configuration.ResourceQOSConfigKey: `
{
  "clusterStrategy": {
    "beClass": {
      "cpuQOS": {
        "groupIdentity": 0
      }
    }
  }
}
`,
				},
			}},
			want: &slov1alpha1.NodeSLOSpec{
				ResourceUsedThresholdWithBE: testingResourceThresholdStrategy,
				ResourceQOSStrategy:         testingResourceQOSStrategy,
				CPUBurstStrategy:            sloconfig.DefaultCPUBurstStrategy(),
				SystemStrategy:              sloconfig.DefaultSystemStrategy(),
				Extensions:                  testingExtensions,
			},
			wantErr: false,
		},
		{
			name: "get spec successfully from old qos config",
			args: args{
				node:    &corev1.Node{},
				nodeSLO: &slov1alpha1.NodeSLO{},
			},
			fields: fields{configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sloconfig.SLOCtrlConfigMap,
					Namespace: sloconfig.ConfigNameSpace,
				},
				Data: map[string]string{
					configuration.ResourceThresholdConfigKey: "{\"clusterStrategy\":{\"enable\":false,\"cpuSuppressThresholdPercent\":60}}",
					configuration.ResourceQOSConfigKey: `
{
  "clusterStrategy": {
    "beClass": {
      "cpuQOS": {
        "groupIdentity": 0
      }
    }
  }
}
`,
				},
			}},
			want: &slov1alpha1.NodeSLOSpec{
				ResourceUsedThresholdWithBE: testingResourceThresholdStrategy,
				ResourceQOSStrategy:         testingResourceQOSStrategyOld,
				CPUBurstStrategy:            sloconfig.DefaultCPUBurstStrategy(),
				SystemStrategy:              sloconfig.DefaultSystemStrategy(),
				Extensions:                  testingExtensions,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctr := &NodeSLOReconciler{Client: fake.NewClientBuilder().WithScheme(scheme).Build()}
			configMapCacheHandler := NewSLOCfgHandlerForConfigMapEvent(ctr.Client, DefaultSLOCfg(), &record.FakeRecorder{})
			ctr.sloCfgCache = configMapCacheHandler
			if tt.fields.configMap != nil {
				ctr.Client.Create(context.Background(), tt.fields.configMap)
				configMapCacheHandler.SyncCacheIfChanged(tt.fields.configMap)
			}

			err := ctr.initNodeSLO(tt.args.node, tt.args.nodeSLO)
			got := &tt.args.nodeSLO.Spec
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeSLOReconciler.initNodeSLO() gotErr = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNodeSLOReconciler_Reconcile(t *testing.T) {
	// initial variants
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	slov1alpha1.AddToScheme(scheme)
	r := &NodeSLOReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
		Scheme: scheme,
	}

	configMapCacheHandler := NewSLOCfgHandlerForConfigMapEvent(r.Client, DefaultSLOCfg(), &record.FakeRecorder{})
	r.sloCfgCache = configMapCacheHandler

	testingNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}
	testingConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sloconfig.SLOCtrlConfigMap,
			Namespace: sloconfig.ConfigNameSpace,
		},
		Data: map[string]string{
			configuration.ResourceThresholdConfigKey: "{\"clusterStrategy\":{\"enable\":true,\"cpuSuppressThresholdPercent\":60}}",
			configuration.ResourceQOSConfigKey: `
{
  "clusterStrategy": {
    "beClass": {
      "cpuQOS": {
        "groupIdentity": 0
      }
    }
  }
}
`,
			configuration.CPUBurstConfigKey: "{\"clusterStrategy\":{\"cfsQuotaBurstPeriodSeconds\":60}}",
			configuration.SystemConfigKey:   "{\"clusterStrategy\":{\"minFreeKbytesFactor\":150,\"watermarkScaleFactor\":150}}",
		},
	}
	testingResourceThresholdStrategy := sloconfig.DefaultResourceThresholdStrategy()
	testingResourceThresholdStrategy.Enable = pointer.Bool(true)
	testingResourceThresholdStrategy.CPUSuppressThresholdPercent = pointer.Int64(60)
	testingResourceQOSStrategy := &slov1alpha1.ResourceQOSStrategy{
		BEClass: &slov1alpha1.ResourceQOS{
			CPUQOS: &slov1alpha1.CPUQOSCfg{
				CPUQOS: slov1alpha1.CPUQOS{
					GroupIdentity: pointer.Int64(0),
				},
			},
		},
	}

	testingCPUBurstStrategy := sloconfig.DefaultCPUBurstStrategy()
	testingCPUBurstStrategy.CFSQuotaBurstPeriodSeconds = pointer.Int64(60)

	testingSystemStrategy := sloconfig.DefaultSystemStrategy()
	testingSystemStrategy.MinFreeKbytesFactor = pointer.Int64(150)

	testingExtensionsMap := *getDefaultExtensionStrategy()
	testingExtensionsIfMap, err := getExtensionsIfMap(testingExtensionsMap)
	if err != nil {
		t.Errorf("failed to get extensions interface map, err:%s", err)
	}

	nodeSLOSpec := &slov1alpha1.NodeSLOSpec{
		ResourceUsedThresholdWithBE: testingResourceThresholdStrategy,
		ResourceQOSStrategy:         testingResourceQOSStrategy,
		CPUBurstStrategy:            testingCPUBurstStrategy,
		SystemStrategy:              testingSystemStrategy,
		Extensions:                  testingExtensionsIfMap,
	}
	nodeReq := ctrl.Request{NamespacedName: types.NamespacedName{Name: testingNode.Name}}
	// the NodeSLO does not exists before getting created
	nodeSLO := &slov1alpha1.NodeSLO{}
	err = r.Client.Get(context.TODO(), nodeReq.NamespacedName, nodeSLO)
	if !errors.IsNotFound(err) {
		t.Errorf("the testing NodeSLO should not exist before getting created, err: %s", err)
	}

	// test cfg not exist, use default config
	result, err := r.Reconcile(context.TODO(), nodeReq)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{Requeue: false}, result, "check_result")
	assert.Equal(t, true, r.sloCfgCache.IsCfgAvailable())

	// throw an error if the configmap does not exist
	err = r.Client.Create(context.TODO(), testingNode)
	assert.NoError(t, err)
	_, err = r.Reconcile(context.TODO(), nodeReq)
	assert.NoError(t, err)
	// create and init a NodeSLO cr if the Node and the configmap exists
	err = r.Client.Create(context.TODO(), testingConfigMap)
	assert.NoError(t, err)
	configMapCacheHandler.SyncCacheIfChanged(testingConfigMap)
	_, err = r.Reconcile(context.TODO(), nodeReq)
	assert.NoError(t, err)
	nodeSLO = &slov1alpha1.NodeSLO{}
	err = r.Client.Get(context.TODO(), nodeReq.NamespacedName, nodeSLO)
	assert.NoError(t, err)
	assert.Equal(t, *nodeSLOSpec, nodeSLO.Spec)
	// delete the NodeSLO cr if the node no longer exists
	err = r.Delete(context.TODO(), testingNode)
	assert.NoError(t, err)
	_, err = r.Reconcile(context.TODO(), nodeReq)
	assert.NoError(t, err)
	nodeSLO = &slov1alpha1.NodeSLO{}
	err = r.Client.Get(context.TODO(), nodeReq.NamespacedName, nodeSLO)
	if !errors.IsNotFound(err) {
		t.Errorf("the testing NodeSLO should not exist after the Node is deleted, err: %s", err)
	}
}
