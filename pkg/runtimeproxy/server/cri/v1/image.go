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

	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
)

var (
	_ v1.ImageServiceServer = (*imageServer)(nil)
)

func NewImageServer(client v1.ImageServiceClient) *imageServer {
	return &imageServer{backendImageServiceClient: client}
}

type imageServer struct {
	backendImageServiceClient v1.ImageServiceClient
}

func (s *imageServer) PullImage(ctx context.Context, req *v1.PullImageRequest) (*v1.PullImageResponse, error) {
	return s.backendImageServiceClient.PullImage(ctx, req)
}

func (s *imageServer) ImageStatus(ctx context.Context, req *v1.ImageStatusRequest) (*v1.ImageStatusResponse, error) {
	return s.backendImageServiceClient.ImageStatus(ctx, req)
}

func (s *imageServer) RemoveImage(ctx context.Context, req *v1.RemoveImageRequest) (*v1.RemoveImageResponse, error) {
	return s.backendImageServiceClient.RemoveImage(ctx, req)
}

func (s *imageServer) ListImages(ctx context.Context, req *v1.ListImagesRequest) (*v1.ListImagesResponse, error) {
	return s.backendImageServiceClient.ListImages(ctx, req)
}

func (s *imageServer) ImageFsInfo(ctx context.Context, req *v1.ImageFsInfoRequest) (*v1.ImageFsInfoResponse, error) {
	return s.backendImageServiceClient.ImageFsInfo(ctx, req)
}
