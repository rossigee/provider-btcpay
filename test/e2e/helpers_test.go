//go:build integration
// +build integration

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

package e2e

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// waitForCondition waits for a resource to have a specific condition.
func waitForCondition(ctx context.Context, c client.Client, obj client.Object, conditionType xpv1.ConditionType, conditionStatus xpv1.ConditionStatus, timeout time.Duration) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	deadline := time.Now().Add(timeout)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for condition %s=%s", conditionType, conditionStatus)
			}
			
			// Get the latest version of the object
			key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
			if err := c.Get(ctx, key, obj); err != nil {
				if client.IgnoreNotFound(err) != nil {
					return err
				}
				// Resource not found, continue waiting
				continue
			}
			
			// Check if the object has the expected condition
			switch v := obj.(type) {
			case interface{ GetCondition(xpv1.ConditionType) xpv1.Condition }:
				if xpv1.IsCondition(v.GetCondition(conditionType), conditionStatus) {
					return nil
				}
			default:
				return fmt.Errorf("unsupported object type: %T", obj)
			}
		}
	}
}

// waitForDeletion waits for a resource to be deleted.
func waitForDeletion(ctx context.Context, c client.Client, obj client.Object, timeout time.Duration) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	deadline := time.Now().Add(timeout)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for deletion")
			}
			
			// Check if the object still exists
			key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
			err := c.Get(ctx, key, obj)
			if err != nil {
				if client.IgnoreNotFound(err) == nil {
					// Resource not found, deletion complete
					return nil
				}
				// Other error
				return err
			}
			// Resource still exists, continue waiting
		}
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}