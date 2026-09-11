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
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/rossigee/provider-btcpay/internal/controller/apikey"
	"github.com/rossigee/provider-btcpay/internal/controller/guest"
	"github.com/rossigee/provider-btcpay/internal/controller/invoicev1beta1"
	"github.com/rossigee/provider-btcpay/internal/controller/providerconfig"
	"github.com/rossigee/provider-btcpay/internal/controller/sharedlink"
	"github.com/rossigee/provider-btcpay/internal/controller/storev1beta1"
	"github.com/rossigee/provider-btcpay/internal/controller/user"
	"github.com/rossigee/provider-btcpay/internal/controller/webhook"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Setup creates all BTCPay controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	if err := providerconfig.Setup(mgr); err != nil {
		return err
	}
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		storev1beta1.Setup,
		invoicev1beta1.Setup,
		webhook.Setup,
		user.Setup,
		apikey.Setup,
		guest.Setup,
		sharedlink.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
