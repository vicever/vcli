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
package main

const (
	unflattenedUsageTemplate = `
{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\
{{define "FormatCommands"}}\
{{range .Commands}}\
{{if not .Hidden}}\
  {{.FullCommand}}{{if .Default}}*{{end}}{{template "FormatCommand" .}}
{{.Help|Wrap 4}}
{{end}}\
{{end}}\
{{end}}\
{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help|Wrap 0}}\
{{end}}\
{{end}}\
{{if .Context.SelectedCommand}}\
usage: {{.App.Name}} {{.Context.SelectedCommand}}{{template "FormatUsage" .Context.SelectedCommand}}
{{else}}\
usage: {{.App.Name}}{{template "FormatUsage" .App}}
{{end}}\
{{if .Context.Flags}}\
Flags:
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.Args}}\
Args:
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand}}\
{{if len .Context.SelectedCommand.Commands}}\
Subcommands:
{{template "FormatCommands" .Context.SelectedCommand}}
{{end}}\
{{else if .App.Commands}}\
Commands:
{{template "FormatCommands" .App}}
{{end}}\
`
)
