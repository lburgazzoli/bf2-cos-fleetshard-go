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

package main

import (
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/camel"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard"
	"os"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"

	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/client"

	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
	utilruntime.Must(cosv2.AddToScheme(fleetshard.Scheme))
	utilruntime.Must(camelv1alpha1.AddToScheme(fleetshard.Scheme))
	utilruntime.Must(camelv1.AddToScheme(fleetshard.Scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	c := controller.Controller{
		Owned:     []client.Object{&camelv1alpha1.KameletBinding{}},
		ApplyFunc: camel.Reconcile,
	}

	if err := fleetshard.Start(c); err != nil {
		ctrl.Log.WithName("setup").Error(err, "problem running manager")
		os.Exit(1)
	}
}
