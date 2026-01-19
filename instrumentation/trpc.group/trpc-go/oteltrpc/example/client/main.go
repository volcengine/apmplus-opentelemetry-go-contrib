// Copyright 2026 Beijing Volcano Engine Technology Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"time"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/log"

	_ "github.com/volcengine/apmplus-opentelemetry-go-contrib/instrumentation/trpc.group/trpc-go/oteltrpc"
	"github.com/volcengine/apmplus-opentelemetry-go-contrib/instrumentation/trpc.group/trpc-go/oteltrpc/example"
	"github.com/volcengine/apmplus-opentelemetry-go-contrib/instrumentation/trpc.group/trpc-go/oteltrpc/example/pb"
)

func main() {
	ctx := context.Background()
	shutdown, err := example.SetupOTelSDK(ctx)
	if err != nil {
		log.Fatalf("failed to setup OTelSDK: %v", err)
		return
	}
	defer func() {
		_ = shutdown(ctx)
	}()

	_ = trpc.NewServer()

	c := pb.NewGreeterClientProxy(client.WithTarget("ip://localhost:8001"))
	for i := 0; i < 1000; i++ {
		message := "hello"
		if i%2 == 0 {
			message = "error"
		}
		rsp, err := c.Hello(context.Background(), &pb.HelloRequest{Msg: message})
		if err != nil {
			log.Error(err)
		} else {
			log.Info(rsp.Msg)
		}
		time.Sleep(time.Second * 10)
	}
}
