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

package v1

import (
	"context"

	"github.com/koordinator-sh/koordinator/pkg/runtimeproxy/server/cri/interceptor"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
)

var (
	_ v1.RuntimeServiceServer = (*runtimeServer)(nil)
)

func NewRuntimeServer(client v1.RuntimeServiceClient, in interceptor.RuntimeInterceptor) *runtimeServer {
	if in == nil {
		in = interceptor.RuntimeNoopInterceptor{}
	}

	return &runtimeServer{backendRuntimeServiceClient: client, interceptor: in}
}

type runtimeServer struct {
	backendRuntimeServiceClient v1.RuntimeServiceClient
	interceptor                 interceptor.RuntimeInterceptor
}

func (s *runtimeServer) Version(ctx context.Context, req *v1.VersionRequest) (*v1.VersionResponse, error) {
	return s.backendRuntimeServiceClient.Version(ctx, req)
}

func (s *runtimeServer) RunPodSandbox(ctx context.Context, req *v1.RunPodSandboxRequest) (*v1.RunPodSandboxResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.RunPodSandbox, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.RunPodSandbox(ctx, req.(*v1.RunPodSandboxRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1.RunPodSandboxResponse), err
}

func (s *runtimeServer) StopPodSandbox(ctx context.Context, req *v1.StopPodSandboxRequest) (*v1.StopPodSandboxResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.StopPodSandbox, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.StopPodSandbox(ctx, req.(*v1.StopPodSandboxRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1.StopPodSandboxResponse), err
}

func (s *runtimeServer) RemovePodSandbox(ctx context.Context, req *v1.RemovePodSandboxRequest) (*v1.RemovePodSandboxResponse, error) {
	return s.backendRuntimeServiceClient.RemovePodSandbox(ctx, req)
}

func (s *runtimeServer) PodSandboxStatus(ctx context.Context, req *v1.PodSandboxStatusRequest) (*v1.PodSandboxStatusResponse, error) {
	return s.backendRuntimeServiceClient.PodSandboxStatus(ctx, req)
}

func (s *runtimeServer) ListPodSandbox(ctx context.Context, req *v1.ListPodSandboxRequest) (*v1.ListPodSandboxResponse, error) {
	return s.backendRuntimeServiceClient.ListPodSandbox(ctx, req)
}

func (s *runtimeServer) CreateContainer(ctx context.Context, req *v1.CreateContainerRequest) (*v1.CreateContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.CreateContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.CreateContainer(ctx, req.(*v1.CreateContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1.CreateContainerResponse), err
}

func (s *runtimeServer) StartContainer(ctx context.Context, req *v1.StartContainerRequest) (*v1.StartContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.StartContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.StartContainer(ctx, req.(*v1.StartContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1.StartContainerResponse), err
}

func (s *runtimeServer) StopContainer(ctx context.Context, req *v1.StopContainerRequest) (*v1.StopContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.StopContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.StopContainer(ctx, req.(*v1.StopContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1.StopContainerResponse), err
}

func (s *runtimeServer) RemoveContainer(ctx context.Context, req *v1.RemoveContainerRequest) (*v1.RemoveContainerResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.RemoveContainer, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.RemoveContainer(ctx, req.(*v1.RemoveContainerRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1.RemoveContainerResponse), err
}

func (s *runtimeServer) ContainerStatus(ctx context.Context, req *v1.ContainerStatusRequest) (*v1.ContainerStatusResponse, error) {
	return s.backendRuntimeServiceClient.ContainerStatus(ctx, req)
}

func (s *runtimeServer) ListContainers(ctx context.Context, req *v1.ListContainersRequest) (*v1.ListContainersResponse, error) {
	return s.backendRuntimeServiceClient.ListContainers(ctx, req)
}

func (s *runtimeServer) UpdateContainerResources(ctx context.Context, req *v1.UpdateContainerResourcesRequest) (*v1.UpdateContainerResourcesResponse, error) {
	rsp, err := s.interceptor.InterceptRuntimeRequest(interceptor.UpdateContainerResources, ctx, req,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.backendRuntimeServiceClient.UpdateContainerResources(ctx, req.(*v1.UpdateContainerResourcesRequest))
		})
	if err != nil {
		return nil, err
	}
	return rsp.(*v1.UpdateContainerResourcesResponse), err
}

func (s *runtimeServer) ContainerStats(ctx context.Context, req *v1.ContainerStatsRequest) (*v1.ContainerStatsResponse, error) {
	return s.backendRuntimeServiceClient.ContainerStats(ctx, req)
}

func (s *runtimeServer) ListContainerStats(ctx context.Context, req *v1.ListContainerStatsRequest) (*v1.ListContainerStatsResponse, error) {
	return s.backendRuntimeServiceClient.ListContainerStats(ctx, req)
}

func (s *runtimeServer) Status(ctx context.Context, req *v1.StatusRequest) (*v1.StatusResponse, error) {
	return s.backendRuntimeServiceClient.Status(ctx, req)
}

func (s *runtimeServer) ReopenContainerLog(ctx context.Context, req *v1.ReopenContainerLogRequest) (*v1.ReopenContainerLogResponse, error) {
	return s.backendRuntimeServiceClient.ReopenContainerLog(ctx, req)
}

func (s *runtimeServer) ExecSync(ctx context.Context, req *v1.ExecSyncRequest) (*v1.ExecSyncResponse, error) {
	return s.backendRuntimeServiceClient.ExecSync(ctx, req)
}

func (s *runtimeServer) Exec(ctx context.Context, req *v1.ExecRequest) (*v1.ExecResponse, error) {
	return s.backendRuntimeServiceClient.Exec(ctx, req)
}

func (s *runtimeServer) Attach(ctx context.Context, req *v1.AttachRequest) (*v1.AttachResponse, error) {
	return s.backendRuntimeServiceClient.Attach(ctx, req)
}

func (s *runtimeServer) PortForward(ctx context.Context, req *v1.PortForwardRequest) (*v1.PortForwardResponse, error) {
	return s.backendRuntimeServiceClient.PortForward(ctx, req)
}

func (s *runtimeServer) UpdateRuntimeConfig(ctx context.Context, req *v1.UpdateRuntimeConfigRequest) (*v1.UpdateRuntimeConfigResponse, error) {
	return s.backendRuntimeServiceClient.UpdateRuntimeConfig(ctx, req)
}

func (s *runtimeServer) PodSandboxStats(ctx context.Context, req *v1.PodSandboxStatsRequest) (*v1.PodSandboxStatsResponse, error) {
	return s.backendRuntimeServiceClient.PodSandboxStats(ctx, req)
}

func (s *runtimeServer) ListPodSandboxStats(ctx context.Context, req *v1.ListPodSandboxStatsRequest) (*v1.ListPodSandboxStatsResponse, error) {
	return s.backendRuntimeServiceClient.ListPodSandboxStats(ctx, req)
}
