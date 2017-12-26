//Copyright 2017 oVirt-maintainers
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package model

type DiskAttachment struct {
	Id          string `json:"id"`
	Bootable    bool   `json:"bootable"`
	PassDiscard bool   `json:"pass_discard"`
	Interface   string `json:"interface"`
	Active      bool   `json:"active"`
	Disk        Disk   `json:"disk"`
}

type Disk struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	ProvisionedSize string `json:"provisioned_size`
}
