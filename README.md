# APMPlus OpenTelemetry Go Contrib

English | [中文README](README.zh_CN.md)

## Introduction

This project is an official Go SDK extension package provided by APMPlus, designed to offer tracing instrumentation for third-party packages.

* Requires: Go 1.22

## Contents

- [Instrumentation](./instrumentation/): Packages providing instrumentation for 3rd-party libraries.

### Supported Libraries
| Go Package                                         | Metrics | Traces |
|----------------------------------------------------|---------| ------ |
| [trpc](instrumentation/trpc.group/trpc-go/oteltrpc) | ✓ | ✓ |

## Security and privacy
This project takes security seriously.
For vulnerability reporting and supported versions, see [SECURITY.md](SECURITY.md).

## License

This project is licensed under the [Apache-2.0 License](LICENSE). 