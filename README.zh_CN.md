# APMPlus OpenTelemetry Go Contrib

中文README | [English](README.md)

## 项目介绍

本项目是由 APMPlus 提供的 Go SDK 扩展包，旨在为第三方包提供追踪埋点能力。

* 要求: Go 1.22

## 目录

- [Instrumentation](./instrumentation/): 为第三方库提供工具的软件包

### 当前支持的库
| Go Package                                         | Metrics | Traces |
|----------------------------------------------------|---------| ------ |
| [trpc](instrumentation/trpc.group/trpc-go/oteltrpc) | ✓ | ✓ |

## 开源协议

本项目采用[Apache-2.0 License](LICENSE)协议.