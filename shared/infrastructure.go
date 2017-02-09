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

const (
	GCPInf    = "google"
	VMWareInf = "vmware"
)

type VMWareInfrastructure struct {
	Type         string `yaml:"type" json:"type"`
	VCenterIP    string `yaml:"vcenter" nav:"vCenter"`
	DataCenter   string `yaml:"datacenter" nav:"datacenter"`
	HostCluster  string `yaml:"hostcluster" nav:"host cluster"`
	Storage      string `yaml:"storagecluster" nav:"storage"`
	ResourcePool string `yaml:"resourcepool" nav:"resource pool"`
}

type GCPInfrastructure struct {
	Type   string `yaml:"type" json:"type"`
	Bucket string `yaml:"bucket" json:"bucket"`
	Zone   string `yaml:"zone" json:"zone"`
	Key    []byte `yaml:"key" json:"key"`
}

type GoogleKey struct {
	Type,
	Project_id,
	Private_key_id,
	Private_key,
	Client_email,
	Client_id,
	Auth_uri,
	Token_uri,
	Auth_provider_x509_cert_url,
	Client_x509_cert_url string
}
