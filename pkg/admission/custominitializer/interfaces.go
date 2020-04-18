/*
Copyright 2017 The Kubernetes Authors.

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

package custominitializer

import (
	"k8s.io/apiserver/pkg/admission"

	informers "github.com/programming-kubernetes/pizza-apiserver/pkg/generated/informers/externalversions"
)

// WantsRestaurantInformerFactory defines a function which sets InformerFactory for admission plugins that need it
// WantsRestaurantInformerFactory定义了函数，它能设置InformerFactory，为需要它的admission plugins
type WantsRestaurantInformerFactory interface {
	SetRestaurantInformerFactory(informers.SharedInformerFactory)
	admission.InitializationValidator
}
