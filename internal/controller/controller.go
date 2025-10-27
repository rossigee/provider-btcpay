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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"

	"github.com/rossigee/provider-btcpay/internal/controller/config"
	"github.com/rossigee/provider-btcpay/internal/controller/invoice"
	"github.com/rossigee/provider-btcpay/internal/controller/paymentmethod"
	"github.com/rossigee/provider-btcpay/internal/controller/store"
	"github.com/rossigee/provider-btcpay/internal/controller/user"
	"github.com/rossigee/provider-btcpay/internal/controller/webhook"
)

// Setup creates all BTCPay controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		config.Setup,
		// v1alpha1 controllers (cluster-scoped)
		store.Setup,
		invoice.Setup,
		user.Setup,
		webhook.Setup,
		paymentmethod.Setup,
		// TODO: Re-add v1beta1 controllers after fixing v2 compatibility
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
