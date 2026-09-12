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

// Package apis contains Kubernetes API for the BTCPay provider.
package apis

import (
	apikeyv1beta1 "github.com/rossigee/provider-btcpay/apis/apikey/v1beta1"
	guestv1beta1 "github.com/rossigee/provider-btcpay/apis/guest/v1beta1"
	invoicev1beta1 "github.com/rossigee/provider-btcpay/apis/invoice/v1beta1"
	sharedlinkv1beta1 "github.com/rossigee/provider-btcpay/apis/sharedlink/v1beta1"
	storev1beta1 "github.com/rossigee/provider-btcpay/apis/store/v1beta1"
	userv1beta1 "github.com/rossigee/provider-btcpay/apis/user/v1beta1"
	"github.com/rossigee/provider-btcpay/apis/v1beta1"
	webhookv1beta1 "github.com/rossigee/provider-btcpay/apis/webhook/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	AddToSchemes = append(AddToSchemes,
		v1beta1.AddToScheme,
		storev1beta1.AddToScheme,
		invoicev1beta1.AddToScheme,
		apikeyv1beta1.AddToScheme,
		guestv1beta1.AddToScheme,
		sharedlinkv1beta1.AddToScheme,
		userv1beta1.AddToScheme,
		webhookv1beta1.AddToScheme,
	)
}

var AddToSchemes runtime.SchemeBuilder

func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
