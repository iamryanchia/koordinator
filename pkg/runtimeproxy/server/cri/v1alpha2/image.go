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

	v1alpha2 "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

var (
	_ v1alpha2.ImageServiceServer = (*imageServer)(nil)
)

func NewImageServer(client v1alpha2.ImageServiceClient) *imageServer {
	return &imageServer{backendImageServiceClient: client}
}

type imageServer struct {
	backendImageServiceClient v1alpha2.ImageServiceClient
}

func (s *imageServer) PullImage(ctx context.Context, req *v1alpha2.PullImageRequest) (*v1alpha2.PullImageResponse, error) {
	return s.backendImageServiceClient.PullImage(ctx, req)
}

func (s *imageServer) ImageStatus(ctx context.Context, req *v1alpha2.ImageStatusRequest) (*v1alpha2.ImageStatusResponse, error) {
	return s.backendImageServiceClient.ImageStatus(ctx, req)
}

func (s *imageServer) RemoveImage(ctx context.Context, req *v1alpha2.RemoveImageRequest) (*v1alpha2.RemoveImageResponse, error) {
	return s.backendImageServiceClient.RemoveImage(ctx, req)
}

func (s *imageServer) ListImages(ctx context.Context, req *v1alpha2.ListImagesRequest) (*v1alpha2.ListImagesResponse, error) {
	return s.backendImageServiceClient.ListImages(ctx, req)
}

func (s *imageServer) ImageFsInfo(ctx context.Context, req *v1alpha2.ImageFsInfoRequest) (*v1alpha2.ImageFsInfoResponse, error) {
	return s.backendImageServiceClient.ImageFsInfo(ctx, req)
}
