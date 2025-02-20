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

package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type RuntimeInterceptor interface {
	InterceptRuntimeRequest(serviceType RuntimeServiceType, ctx context.Context,
		req interface{}, handler grpc.UnaryHandler) (interface{}, error)
}

type RuntimeNoopInterceptor struct{}

func (i RuntimeNoopInterceptor) InterceptRuntimeRequest(serviceType RuntimeServiceType, ctx context.Context,
	req interface{}, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
