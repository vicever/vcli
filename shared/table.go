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

import (
	"errors"
	"os"

	"github.com/sisatech/tablewriter"
)

func PrettyTable(vals [][]string) {

	if len(vals) == 0 {
		panic(errors.New("no rows provided"))
	}

	// produce table for output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.UseArcBoxBorders()
	table.SetHeader(vals[0])
	for i := 1; i < len(vals); i++ {
		table.Append(vals[i])
	}
	table.Render()

}

func PrettyLeftTable(vals [][]string) {

	if len(vals) == 0 {
		panic(errors.New("no rows provided"))
	}

	// produce table for output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.UseArcBoxBorders()
	table.SetHeader(vals[0])
	for i := 1; i < len(vals); i++ {
		table.Append(vals[i])
	}
	table.Render()

}
