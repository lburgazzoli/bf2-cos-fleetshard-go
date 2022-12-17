/*
Copyright 2022.

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

package cos

import (
	"encoding/json"
	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patch(oldResource client.Object, newResource client.Object) ([]byte, error) {

	oldJson, err := json.Marshal(oldResource)
	if err != nil {
		return nil, err
	}
	newJson, err := json.Marshal(newResource)
	if err != nil {
		return nil, err
	}

	// NOTE: this is likely not correct
	return strategicpatch.CreateTwoWayMergePatch(oldJson, newJson, cos.ManagedConnector{})
}
