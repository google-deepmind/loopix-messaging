// Copyright 2018 The Loopix-Messaging Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sphinx

import "fmt"

func XorBytes(b1, b2 []byte) []byte {

	if len(b1) != len(b2) {
		panic("String cannot be xored if their length is different")
	}

	b := make([]byte, len(b1))
	for i, _ := range b {
		b[i] = b1[i] ^ b2[i]
	}
	return b
}

func BytesToString(b []byte) string {
	result := ""
	for _, v := range b {
		s := fmt.Sprintf("%v", v)
		result = result + s
	}
	return result
}
