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

package filters

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	trpcpb "trpc.group/trpc/trpc-protocol/pb/go/trpc"
)

const (
	scopeName    = "trpc.group/trpc-go/trpc-go"
	scopeVersion = "1.0.0"
)

const (
	NamespaceKey     = attribute.Key("rpc.trpc.namespace")
	EnvNameKey       = attribute.Key("rpc.trpc.env_name")
	CallerServiceKey = attribute.Key("rpc.trpc.caller_service")
	CalleeServiceKey = attribute.Key("rpc.trpc.callee_service")
	CallerMethodKey  = attribute.Key("rpc.trpc.caller_method")
	CalleeMethodKey  = attribute.Key("rpc.trpc.callee_method")
	ErrorMessageKey  = attribute.Key("error.message")

	system = "trpc"
)

var (
	defaultTracerOnce sync.Once

	defaultTracer           trace.Tracer
	clientDurationHistogram metric.Int64Histogram
	serverDurationHistogram metric.Int64Histogram
)

func getDefaultTracer() trace.Tracer {
	defaultTracerOnce.Do(func() {
		defaultTracer = otel.Tracer(scopeName, trace.WithInstrumentationVersion(scopeVersion))

		meter := otel.Meter(scopeName, metric.WithInstrumentationVersion(scopeVersion))
		clientDurationHistogram, _ = meter.Int64Histogram("rpc.client.duration", metric.WithUnit("ms"), metric.WithDescription("The duration of an outbound RPC invocation."))
		serverDurationHistogram, _ = meter.Int64Histogram("rpc.server.duration", metric.WithUnit("ms"), metric.WithDescription("The duration of an inbound RPC invocation."))
	})
	return defaultTracer
}

// ServerFilter is an apmplus filter for server.
func ServerFilter(ctx context.Context, req interface{}, next filter.ServerHandleFunc) (rsp interface{}, err error) {
	var (
		msg        = trpc.Message(ctx)
		md         = msg.ServerMetaData()
		attrs      = buildCommonAttributes(msg.ServerRPCName())
		afterAttrs []attribute.KeyValue
		start      = time.Now()
	)

	if md == nil {
		md = codec.MetaData{}
	}

	ctx = otel.GetTextMapPropagator().Extract(ctx, newMetadataSupplier(md))
	ctx, span := getDefaultTracer().Start(
		trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(ctx)),
		msg.ServerRPCName(),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...),
		trace.WithAttributes(CallerServiceKey.String(msg.CallerServiceName())),
		trace.WithAttributes(CallerMethodKey.String(msg.CallerMethod())),
		trace.WithAttributes(NamespaceKey.String(msg.Namespace())),
		trace.WithAttributes(EnvNameKey.String(msg.EnvName())))

	defer func() {
		span.SetAttributes(afterAttrs...)
		span.End()
		serverDurationHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attrs...), metric.WithAttributes(afterAttrs...))
	}()

	rsp, err = next(ctx, req)
	afterAttrs = handleError(err, span)
	afterAttrs = append(afterAttrs, netAttributes(addressString(msg.LocalAddr()), addressString(msg.RemoteAddr()))...)
	return rsp, err
}

// ClientFilter is an apmplus filter for client.
func ClientFilter(ctx context.Context, req, rsp interface{}, next filter.ClientHandleFunc) error {
	var (
		msg        = trpc.Message(ctx)
		md         = msg.ClientMetaData()
		attrs      = buildCommonAttributes(msg.ClientRPCName())
		afterAttrs = make([]attribute.KeyValue, 0, 4)
		opts       = client.OptionsFromContext(ctx)
		start      = time.Now()
	)

	if md == nil {
		md = codec.MetaData{}
	}
	ctx, span := getDefaultTracer().Start(ctx,
		msg.ClientRPCName(),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
		trace.WithAttributes(CalleeServiceKey.String(msg.CalleeServiceName())),
		trace.WithAttributes(CalleeMethodKey.String(msg.CalleeMethod())),
		trace.WithAttributes(NamespaceKey.String(msg.Namespace())),
		trace.WithAttributes(EnvNameKey.String(msg.EnvName())),
	)
	defer func() {
		span.SetAttributes(afterAttrs...)
		span.End()
		clientDurationHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attrs...), metric.WithAttributes(afterAttrs...))
	}()

	otel.GetTextMapPropagator().Inject(ctx, newMetadataSupplier(md))
	msg.WithClientMetaData(md)

	err := next(ctx, req, rsp)
	afterAttrs = append(afterAttrs, handleError(err, span)...)
	afterAttrs = append(afterAttrs, netAttributes(parseTarget(opts.Target), addressString(msg.RemoteAddr()))...)
	return err
}

func splitServiceAndMethod(s string) (string, string) {
	if s == "" {
		return "", ""
	}
	idx := strings.LastIndexByte(s, '/')
	if idx <= 0 {
		return s, ""
	}
	return strings.Trim(s[:idx], "/"), s[idx+1:]
}

func handleError(err error, span trace.Span) []attribute.KeyValue {
	if err == nil {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	span.RecordError(err)
	code, msg := getErrCode(err)
	span.SetStatus(codes.Error, msg)
	return []attribute.KeyValue{
		semconv.ErrorTypeKey.Int(code),
		ErrorMessageKey.String(msg),
	}
}

func getErrCode(err error) (int, string) {
	var trpcErr *errs.Error
	if errors.As(err, &trpcErr) {
		return int(trpcErr.Code), trpcErr.Msg
	}
	return int(trpcpb.TrpcRetCode_TRPC_INVOKE_UNKNOWN_ERR), ""
}

func buildCommonAttributes(rpcName string) []attribute.KeyValue {
	service, method := splitServiceAndMethod(rpcName)
	return []attribute.KeyValue{
		semconv.RPCSystemKey.String(system),
		semconv.RPCServiceKey.String(service),
		semconv.RPCMethodKey.String(method),
	}
}

func netAttributes(serverAddr, remoteAddr string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 4)
	attrs = append(attrs, addressAttributes(semconv.ServerAddressKey, semconv.ServerPortKey, serverAddr)...)
	attrs = append(attrs, addressAttributes(semconv.NetworkPeerAddressKey, semconv.NetworkPeerPortKey, remoteAddr)...)
	return attrs
}

func parseTarget(target string) string {
	if target == "" {
		return ""
	}
	substr := "://"
	idx := strings.Index(target, substr)
	if idx == -1 {
		return ""
	}
	return target[idx+len(substr):]
}

func addressString(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	return addr.String()
}

func addressAttributes(hostKey, portKey attribute.Key, addr string) []attribute.KeyValue {
	if addr == "" {
		return nil
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return []attribute.KeyValue{hostKey.String(host)}
	}
	return []attribute.KeyValue{hostKey.String(host), portKey.Int(p)}
}
