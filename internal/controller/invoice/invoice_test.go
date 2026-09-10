/*
Copyright 2023 The Crossplane Authors.

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

package invoice

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	invoicev1alpha1 "github.com/rossigee/provider-btcpay/apis/invoice/v1alpha1"
)

var _ resource.Managed = &invoicev1alpha1.Invoice{}

func TestInvoiceImplementsManagedInterface(t *testing.T) {
	var _ resource.Managed = &invoicev1alpha1.Invoice{}
}

func TestInvoiceSpecFields(t *testing.T) {
	cr := &invoicev1alpha1.Invoice{}
	_ = cr.Spec.ForProvider.StoreRef
	_ = cr.Spec.ForProvider.Amount
	_ = cr.Spec.ForProvider.Currency
}

func TestInvoiceStatusFields(t *testing.T) {
	cr := &invoicev1alpha1.Invoice{}
	_ = cr.Status.AtProvider.ID
	_ = cr.Status.AtProvider.StoreID
	_ = cr.Status.AtProvider.Amount
	_ = cr.Status.AtProvider.Currency
	_ = cr.Status.AtProvider.Status
}
