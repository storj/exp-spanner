/*
Copyright 2024 Storj Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spanner

import (
	"bytes"
	"encoding/json"
)

type jsoniter_Config struct {
	EscapeHTML             bool
	SortMapKeys            bool
	UseNumber              bool
	ValidateJsonRawMessage bool
}

var jsoniter_ConfigCompatibleWithStandardLibrary = jsoniter_Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}

func (cfg jsoniter_Config) Froze() jsoniter_Config { return cfg }

func (cfg jsoniter_Config) Marshal(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(cfg.EscapeHTML)
	err := enc.Encode(v)
	return b.Bytes(), err
}

func (cfg jsoniter_Config) Unmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if cfg.UseNumber {
		dec.UseNumber()
	}
	return dec.Decode(v)
}
