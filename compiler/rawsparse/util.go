// Copyright 2016 Sisa-Tech Pty Ltd
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
package rawsparse

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

func ceiling(x, y uint64) uint64 {

	return (x + y - 1) / y

}

func generateDiskUID() string {

	rand.Seed(time.Now().UTC().UnixNano())
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], uint32(rand.Int31()))
	return fmt.Sprintf("%X", b)

}
