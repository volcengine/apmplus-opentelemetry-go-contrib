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
	"testing"

	"github.com/stretchr/testify/assert"
	"trpc.group/trpc-go/trpc-go/codec"
)

func Test_metadataSupplier(t *testing.T) {
	supplier := newMetadataSupplier(codec.MetaData{})
	supplier.Set("key", "value")
	assert.Equal(t, "value", supplier.Get("key"))
	assert.Equal(t, []string{"key"}, supplier.Keys())
}
