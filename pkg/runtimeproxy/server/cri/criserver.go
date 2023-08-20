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
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	v1alpha2 "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"

	"github.com/koordinator-sh/koordinator/cmd/koord-runtime-proxy/options"
	"github.com/koordinator-sh/koordinator/pkg/runtimeproxy/config"
	"github.com/koordinator-sh/koordinator/pkg/runtimeproxy/dispatcher"
	resource_executor "github.com/koordinator-sh/koordinator/pkg/runtimeproxy/resexecutor"
	cri_resource_executor "github.com/koordinator-sh/koordinator/pkg/runtimeproxy/resexecutor/cri"
	"github.com/koordinator-sh/koordinator/pkg/runtimeproxy/server/cri/interceptor"
	proxyv1 "github.com/koordinator-sh/koordinator/pkg/runtimeproxy/server/cri/v1"
	proxyv1alpha2 "github.com/koordinator-sh/koordinator/pkg/runtimeproxy/server/cri/v1alpha2"
	"github.com/koordinator-sh/koordinator/pkg/runtimeproxy/utils"
)

const (
	defaultTimeout = 5 * time.Second
)

type RuntimeManagerCriServer struct {
	hookDispatcher *dispatcher.RuntimeHookDispatcher

	runtimeServerV1 v1.RuntimeServiceServer
	imageServerV1   v1.ImageServiceServer

	runtimeServerV1alpha2 v1alpha2.RuntimeServiceServer
	imageServerV1alpha2   v1alpha2.ImageServiceServer
}

func NewRuntimeManagerCriServer() *RuntimeManagerCriServer {
	criInterceptor := &RuntimeManagerCriServer{
		hookDispatcher: dispatcher.NewRuntimeDispatcher(),
	}
	return criInterceptor
}

func (c *RuntimeManagerCriServer) Name() string {
	return "RuntimeManagerCriServer"
}

func (c *RuntimeManagerCriServer) Run() error {
	if err := c.initBackendServer(options.RemoteRuntimeServiceEndpoint, options.RemoteImageServiceEndpoint); err != nil {
		return err
	}
	c.failOver()

	klog.Infof("do failOver done")

	listener, err := net.Listen("unix", options.RuntimeProxyEndpoint)
	if err != nil {
		klog.Errorf("failed to create listener, error: %v", err)
		return err
	}
	grpcServer := grpc.NewServer()

	if c.runtimeServerV1 != nil {
		v1.RegisterRuntimeServiceServer(grpcServer, c.runtimeServerV1)
	}
	if c.imageServerV1 != nil {
		v1.RegisterImageServiceServer(grpcServer, c.imageServerV1)
	}

	if c.runtimeServerV1alpha2 != nil {
		v1alpha2.RegisterRuntimeServiceServer(grpcServer, c.runtimeServerV1alpha2)
	}
	if c.imageServerV1alpha2 != nil {
		v1alpha2.RegisterImageServiceServer(grpcServer, c.imageServerV1alpha2)
	}

	err = grpcServer.Serve(listener)
	return err
}

func (c *RuntimeManagerCriServer) getRuntimeHookInfo(serviceType interceptor.RuntimeServiceType) (config.RuntimeRequestPath,
	resource_executor.RuntimeResourceType) {
	switch serviceType {
	case interceptor.RunPodSandbox:
		return config.RunPodSandbox, resource_executor.RuntimePodResource
	case interceptor.StopPodSandbox:
		return config.StopPodSandbox, resource_executor.RuntimePodResource
	case interceptor.CreateContainer:
		return config.CreateContainer, resource_executor.RuntimeContainerResource
	case interceptor.StartContainer:
		return config.StartContainer, resource_executor.RuntimeContainerResource
	case interceptor.StopContainer:
		return config.StopContainer, resource_executor.RuntimeContainerResource
	case interceptor.UpdateContainerResources:
		return config.UpdateContainerResources, resource_executor.RuntimeContainerResource
	}
	return config.NoneRuntimeHookPath, resource_executor.RuntimeNoopResource
}

func (c *RuntimeManagerCriServer) InterceptRuntimeRequest(serviceType interceptor.RuntimeServiceType,
	ctx context.Context, request interface{}, handler grpc.UnaryHandler) (interface{}, error) {
	runtimeHookPath, runtimeResourceType := c.getRuntimeHookInfo(serviceType)
	resourceExecutor := resource_executor.NewRuntimeResourceExecutor(runtimeResourceType)

	callHookOperation, err := resourceExecutor.ParseRequest(request)
	if err != nil {
		klog.Errorf("fail to parse request %v %v", request, err)
	}
	defer resourceExecutor.DeleteCheckpointIfNeed(request)

	switch callHookOperation {
	case utils.ShouldCallHookPlugin:
		// TODO deal with the Dispatch response
		response, err, policy := c.hookDispatcher.Dispatch(ctx, runtimeHookPath, config.PreHook, resourceExecutor.GenerateHookRequest())
		if err != nil {
			klog.Errorf("fail to call hook server %v", err)
			if policy == config.PolicyFail {
				return nil, fmt.Errorf("hook server err: %v", err)
			}
		} else if response != nil {
			if err = resourceExecutor.UpdateRequest(response, request); err != nil {
				klog.Errorf("failed to update cri request %v", err)
			}
		}
	}
	// call the backend runtime engine
	res, err := handler(ctx, request)
	if err == nil {
		klog.Infof("%v call containerd %v success", resourceExecutor.GetMetaInfo(), string(runtimeHookPath))
		// store checkpoint info basing request only when response success
		if err := resourceExecutor.ResourceCheckPoint(res); err != nil {
			klog.Errorf("fail to checkpoint %v %v", resourceExecutor.GetMetaInfo(), err)
		}
	} else {
		klog.Errorf("%v call containerd %v fail %v", resourceExecutor.GetMetaInfo(), string(runtimeHookPath), err)
	}
	switch callHookOperation {
	case utils.ShouldCallHookPlugin:
		// post call hook server
		// TODO the response
		c.hookDispatcher.Dispatch(ctx, runtimeHookPath, config.PostHook, resourceExecutor.GenerateHookRequest())
	}
	return res, err
}

// initBackendServerV1 initializes runtimeServerV1 and imageServerV1 if the backend server supports CRI v1 API.
// Other methods can compare the two fields with nil to know whether the CRI v1 API is supported.
func (c *RuntimeManagerCriServer) initBackendServerV1(runtimeConn, imageConn *grpc.ClientConn) error {
	runtimeClient := v1.NewRuntimeServiceClient(runtimeConn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := runtimeClient.Version(ctx, &v1.VersionRequest{}); err == nil {
		c.runtimeServerV1 = proxyv1.NewRuntimeServer(runtimeClient, c)
		klog.Infof("the backend runtime server supports CRI v1 API")
	} else if status.Code(err) == codes.Unimplemented {
		klog.Infof("the backend runtime server doesn't support CRI v1 API")
	} else {
		return err
	}

	imageClient := v1.NewImageServiceClient(imageConn)
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := imageClient.ImageFsInfo(ctx, &v1.ImageFsInfoRequest{}); err == nil {
		c.imageServerV1 = proxyv1.NewImageServer(imageClient)
		klog.Infof("the backend image server supports CRI v1 API")
	} else if status.Code(err) == codes.Unimplemented {
		klog.Infof("the backend image server doesn't support CRI v1 API")
	} else {
		return err
	}

	return nil
}

// initBackendServerV1alpha2 initializes runtimeServerV1alpha2 and imageServerV1alpha2 if the backend server supports CRI v1alpha2 API.
// Other methods can compare the two fields with nil to know whether the CRI v1alpha2 API is supported.
func (c *RuntimeManagerCriServer) initBackendServerV1alpha2(runtimeConn, imageConn *grpc.ClientConn) error {
	runtimeClient := v1alpha2.NewRuntimeServiceClient(runtimeConn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := runtimeClient.Version(ctx, &v1alpha2.VersionRequest{}); err == nil {
		c.runtimeServerV1alpha2 = proxyv1alpha2.NewRuntimeServer(runtimeClient, c)
		klog.Infof("the backend runtime server supports CRI v1alpha2 API")
	} else if status.Code(err) == codes.Unimplemented {
		klog.Infof("the backend runtime server doesn't support CRI v1alpha2 API")
	} else {
		return err
	}

	imageClient := v1alpha2.NewImageServiceClient(imageConn)
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := imageClient.ImageFsInfo(ctx, &v1alpha2.ImageFsInfoRequest{}); err == nil {
		c.imageServerV1alpha2 = proxyv1alpha2.NewImageServer(imageClient)
		klog.Infof("the backend image server supports CRI v1alpha2 API")
	} else if status.Code(err) == codes.Unimplemented {
		klog.Infof("the backend image server doesn't support CRI v1alpha2 API")
	} else {
		return err
	}

	return nil
}

func dialer(ctx context.Context, addr string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, "unix", addr)
}

func (c *RuntimeManagerCriServer) initBackendServer(runtimeSockPath, imageSockPath string) error {
	generateGrpcConn := func(sockPath string) (*grpc.ClientConn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		return grpc.DialContext(ctx, sockPath, grpc.WithInsecure(), grpc.WithContextDialer(dialer))
	}

	runtimeConn, err := generateGrpcConn(runtimeSockPath)
	if err != nil {
		klog.Errorf("fail to create runtime service client %v", err)
		return err
	}

	imageConn, err := generateGrpcConn(imageSockPath)
	if err != nil {
		klog.Errorf("fail to create image service client %v", err)
		return err
	}

	if err := c.initBackendServerV1(runtimeConn, imageConn); err != nil {
		klog.Errorf("fail to init CRI v1 server %v", err)
		return err
	}

	if err := c.initBackendServerV1alpha2(runtimeConn, imageConn); err != nil {
		klog.Errorf("fail to init CRI v1alpha2 server %v", err)
		return err
	}

	if c.runtimeServerV1 == nil && c.runtimeServerV1alpha2 == nil {
		return errors.New("the backend runtime server doesn't support CRI v1 or v1alpha2 API")
	}

	if c.imageServerV1 == nil && c.imageServerV1alpha2 == nil {
		return errors.New("the backend image server doesn't support CRI v1 or v1alpha2 API")
	}

	klog.Info("success to init backend servers")
	return nil
}

// listPodsAndContainersV1 lists the current pods and containers through CRI v1 API.
// It's caller's responsibility to ensure the CRI v1 API is supported.
func (c *RuntimeManagerCriServer) listPodsAndContainersV1() (*v1.ListPodSandboxResponse, *v1.ListContainersResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	podResponse, err := c.runtimeServerV1.ListPodSandbox(ctx, &v1.ListPodSandboxRequest{})
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	containerResponse, err := c.runtimeServerV1.ListContainers(ctx, &v1.ListContainersRequest{})
	if err != nil {
		return nil, nil, err
	}

	return podResponse, containerResponse, nil
}

// listPodsAndContainersV1alpha2 lists the current pods and containers through CRI v1alpha2 API,
// and converts them to CRI v1 types to eliminate the verbosity of interacting with multi-version types.
// It's caller's responsibility to ensure the CRI v1alpha2 API is supported.
func (c *RuntimeManagerCriServer) listPodsAndContainersV1alpha2() (*v1.ListPodSandboxResponse, *v1.ListContainersResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	podResponse, err := c.runtimeServerV1alpha2.ListPodSandbox(ctx, &v1alpha2.ListPodSandboxRequest{})
	if err != nil {
		return nil, nil, err
	}

	podResponseV1, err := cri_resource_executor.V1alpha2ToV1(podResponse)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	containerResponse, err := c.runtimeServerV1alpha2.ListContainers(ctx, &v1alpha2.ListContainersRequest{})
	if err != nil {
		return nil, nil, err
	}

	containerResponseV1, err := cri_resource_executor.V1alpha2ToV1(containerResponse)
	if err != nil {
		return nil, nil, err
	}

	return podResponseV1.(*v1.ListPodSandboxResponse), containerResponseV1.(*v1.ListContainersResponse), nil
}

func (c *RuntimeManagerCriServer) failOver() error {
	var podResponse *v1.ListPodSandboxResponse
	var containerResponse *v1.ListContainersResponse
	var err error

	if c.runtimeServerV1 != nil {
		podResponse, containerResponse, err = c.listPodsAndContainersV1()
	} else {
		podResponse, containerResponse, err = c.listPodsAndContainersV1alpha2()
	}
	if err != nil {
		klog.Errorf("failed to list pods and containers %v", err)
		return err
	}

	for _, pod := range podResponse.Items {
		podResourceExecutor := cri_resource_executor.NewPodResourceExecutor()
		if err := podResourceExecutor.ParsePod(pod); err != nil {
			klog.Errorf("failed to parse pod %s, err: %v", pod.Id, err)
			continue
		}
		podResourceExecutor.ResourceCheckPoint(&v1.RunPodSandboxResponse{
			PodSandboxId: pod.GetId(),
		})
	}

	for _, container := range containerResponse.Containers {
		containerExecutor := cri_resource_executor.NewContainerResourceExecutor()
		if err := containerExecutor.ParseContainer(container); err != nil {
			klog.Errorf("failed to parse container %s, err: %v", container.Id, err)
			continue
		}
		containerExecutor.ResourceCheckPoint(&v1.CreateContainerResponse{
			ContainerId: container.GetId(),
		})
	}

	return nil
}
