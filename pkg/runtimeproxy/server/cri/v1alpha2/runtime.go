/*
Copyright 2023 The Koordinator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except req compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to req writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"context"

	"github.com/koordinator-sh/koordinator/pkg/runtimeproxy/server/cri/interceptor"
	v1alpha2 "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

var (
	_ v1alpha2.RuntimeServiceServer = (*runtimeServer)(nil)
)

func NewRuntimeServer(client v1alpha2.RuntimeServiceClient, in interceptor.RuntimeInterceptor) *runtimeServer {
	if in == nil {
		in = interceptor.RuntimeNoopInterceptor{}
	}

	return &runtimeServer{backendRuntimeServiceClient: client, interceptor: in}
}

type runtimeServer struct {
	backendRuntimeServiceClient v1alpha2.RuntimeServiceClient
	interceptor                 interceptor.RuntimeInterceptor
}

func (s *runtimeServer) Version(ctx context.Context, req *v1alpha2.VersionRequest) (*v1alpha2.VersionResponse, error) {
	return s.backendRuntimeServiceClient.Version(ctx, req)
}

func (s *runtimeServer) RunPodSandbox(ctx context.Context, req *v1alpha2.RunPodSandboxRequest) (*v1alpha2.RunPodSandboxResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.RunPodSandbox, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.RunPodSandbox(ctx, req.(*v1alpha2.RunPodSandboxRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1alpha2.RunPodSandboxResponse), err
}

func (s *runtimeServer) StopPodSandbox(ctx context.Context, req *v1alpha2.StopPodSandboxRequest) (*v1alpha2.StopPodSandboxResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.StopPodSandbox, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.StopPodSandbox(ctx, req.(*v1alpha2.StopPodSandboxRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1alpha2.StopPodSandboxResponse), err
}

func (s *runtimeServer) RemovePodSandbox(ctx context.Context, req *v1alpha2.RemovePodSandboxRequest) (*v1alpha2.RemovePodSandboxResponse, error) {
	return s.backendRuntimeServiceClient.RemovePodSandbox(ctx, req)
}

func (s *runtimeServer) PodSandboxStatus(ctx context.Context, req *v1alpha2.PodSandboxStatusRequest) (*v1alpha2.PodSandboxStatusResponse, error) {
	return s.backendRuntimeServiceClient.PodSandboxStatus(ctx, req)
}

func (s *runtimeServer) ListPodSandbox(ctx context.Context, req *v1alpha2.ListPodSandboxRequest) (*v1alpha2.ListPodSandboxResponse, error) {
	return s.backendRuntimeServiceClient.ListPodSandbox(ctx, req)
}

func (s *runtimeServer) CreateContainer(ctx context.Context, req *v1alpha2.CreateContainerRequest) (*v1alpha2.CreateContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.CreateContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.CreateContainer(ctx, req.(*v1alpha2.CreateContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1alpha2.CreateContainerResponse), err
}

func (s *runtimeServer) StartContainer(ctx context.Context, req *v1alpha2.StartContainerRequest) (*v1alpha2.StartContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.StartContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.StartContainer(ctx, req.(*v1alpha2.StartContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1alpha2.StartContainerResponse), err
}

func (s *runtimeServer) StopContainer(ctx context.Context, req *v1alpha2.StopContainerRequest) (*v1alpha2.StopContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.StopContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.StopContainer(ctx, req.(*v1alpha2.StopContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1alpha2.StopContainerResponse), err
}

func (s *runtimeServer) RemoveContainer(ctx context.Context, req *v1alpha2.RemoveContainerRequest) (*v1alpha2.RemoveContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.RemoveContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.RemoveContainer(ctx, req.(*v1alpha2.RemoveContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1alpha2.RemoveContainerResponse), err
}

func (s *runtimeServer) ContainerStatus(ctx context.Context, req *v1alpha2.ContainerStatusRequest) (*v1alpha2.ContainerStatusResponse, error) {
	return s.backendRuntimeServiceClient.ContainerStatus(ctx, req)
}

func (s *runtimeServer) ListContainers(ctx context.Context, req *v1alpha2.ListContainersRequest) (*v1alpha2.ListContainersResponse, error) {
	return s.backendRuntimeServiceClient.ListContainers(ctx, req)
}

func (s *runtimeServer) UpdateContainerResources(ctx context.Context, req *v1alpha2.UpdateContainerResourcesRequest) (*v1alpha2.UpdateContainerResourcesResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.UpdateContainerResources, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.UpdateContainerResources(ctx, req.(*v1alpha2.UpdateContainerResourcesRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1alpha2.UpdateContainerResourcesResponse), err
}

func (s *runtimeServer) ContainerStats(ctx context.Context, req *v1alpha2.ContainerStatsRequest) (*v1alpha2.ContainerStatsResponse, error) {
	return s.backendRuntimeServiceClient.ContainerStats(ctx, req)
}

func (s *runtimeServer) ListContainerStats(ctx context.Context, req *v1alpha2.ListContainerStatsRequest) (*v1alpha2.ListContainerStatsResponse, error) {
	return s.backendRuntimeServiceClient.ListContainerStats(ctx, req)
}

func (s *runtimeServer) Status(ctx context.Context, req *v1alpha2.StatusRequest) (*v1alpha2.StatusResponse, error) {
	return s.backendRuntimeServiceClient.Status(ctx, req)
}

func (s *runtimeServer) ReopenContainerLog(ctx context.Context, req *v1alpha2.ReopenContainerLogRequest) (*v1alpha2.ReopenContainerLogResponse, error) {
	return s.backendRuntimeServiceClient.ReopenContainerLog(ctx, req)
}

func (s *runtimeServer) ExecSync(ctx context.Context, req *v1alpha2.ExecSyncRequest) (*v1alpha2.ExecSyncResponse, error) {
	return s.backendRuntimeServiceClient.ExecSync(ctx, req)
}

func (s *runtimeServer) Exec(ctx context.Context, req *v1alpha2.ExecRequest) (*v1alpha2.ExecResponse, error) {
	return s.backendRuntimeServiceClient.Exec(ctx, req)
}

func (s *runtimeServer) Attach(ctx context.Context, req *v1alpha2.AttachRequest) (*v1alpha2.AttachResponse, error) {
	return s.backendRuntimeServiceClient.Attach(ctx, req)
}

func (s *runtimeServer) PortForward(ctx context.Context, req *v1alpha2.PortForwardRequest) (*v1alpha2.PortForwardResponse, error) {
	return s.backendRuntimeServiceClient.PortForward(ctx, req)
}

func (s *runtimeServer) UpdateRuntimeConfig(ctx context.Context, req *v1alpha2.UpdateRuntimeConfigRequest) (*v1alpha2.UpdateRuntimeConfigResponse, error) {
	return s.backendRuntimeServiceClient.UpdateRuntimeConfig(ctx, req)
}

func (s *runtimeServer) PodSandboxStats(ctx context.Context, req *v1alpha2.PodSandboxStatsRequest) (*v1alpha2.PodSandboxStatsResponse, error) {
	return s.backendRuntimeServiceClient.PodSandboxStats(ctx, req)
}

func (s *runtimeServer) ListPodSandboxStats(ctx context.Context, req *v1alpha2.ListPodSandboxStatsRequest) (*v1alpha2.ListPodSandboxStatsResponse, error) {
	return s.backendRuntimeServiceClient.ListPodSandboxStats(ctx, req)
}
