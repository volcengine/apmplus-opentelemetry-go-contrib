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

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/log"
	trpcpb "trpc.group/trpc/trpc-protocol/pb/go/trpc"

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
	s := trpc.NewServer()
	pb.RegisterGreeterService(s, &Greeter{})
	if err := s.Serve(); err != nil {
		log.Error(err)
	}
}

type Greeter struct{}

func (g Greeter) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Infof("got hello request: %s", req.Msg)
	if req.Msg == "error" {
		return nil, errs.New(trpcpb.TrpcRetCode_TRPC_SERVER_SYSTEM_ERR, "hello error")
	}
	return &pb.HelloReply{Msg: "Hello " + req.Msg + "!"}, nil
}
