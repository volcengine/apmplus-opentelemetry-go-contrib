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
	"trpc.group/trpc-go/trpc-go/codec"
)

type metadataSupplier struct {
	metadata codec.MetaData
}

func newMetadataSupplier(md codec.MetaData) *metadataSupplier {
	return &metadataSupplier{
		metadata: md,
	}
}

func (s *metadataSupplier) Get(key string) string {
	value := s.metadata[key]
	if len(value) == 0 {
		return ""
	}
	return string(value)
}

func (s *metadataSupplier) Set(key string, value string) {
	s.metadata[key] = []byte(value)
}

func (s *metadataSupplier) Keys() []string {
	keys := make([]string, 0, len(s.metadata))
	for key := range s.metadata {
		keys = append(keys, key)
	}
	return keys
}
