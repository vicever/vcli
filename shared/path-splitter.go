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
package shared

import "strings"

// NodeName returns the end-name of a node within a path.
func NodeName(path string) string {

	nodes := strings.Split(path, "/")
	return nodes[len(nodes)-1]

}

// NodeParent returns the path to the parent parent node.
func NodeParent(path string) string {

	nodes := strings.Split(path, "/")

	if len(nodes) == 1 {

		return ""

	}

	return strings.TrimSuffix(path, nodes[len(nodes)-1])

}
