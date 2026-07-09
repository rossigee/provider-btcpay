/*
Copyright 2025 The Crossplane Authors.

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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Invoice type metadata.
var (
	InvoiceKind             = reflect.TypeOf(Invoice{}).Name()
	InvoiceGroupKind        = schema.GroupKind{Group: Group, Kind: InvoiceKind}
	InvoiceKindAPIVersion   = InvoiceKind + "." + SchemeGroupVersion.String()
	InvoiceGroupVersionKind = SchemeGroupVersion.WithKind(InvoiceKind)
)

// StoreReference type metadata.
var (
	StoreReferenceKind             = reflect.TypeOf(StoreReference{}).Name()
	StoreReferenceGroupKind        = schema.GroupKind{Group: Group, Kind: StoreReferenceKind}
	StoreReferenceKindAPIVersion   = StoreReferenceKind + "." + SchemeGroupVersion.String()
	StoreReferenceGroupVersionKind = SchemeGroupVersion.WithKind(StoreReferenceKind)
)
