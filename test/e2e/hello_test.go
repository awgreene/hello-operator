// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	hellov1alpha1 "github.com/awgreene/hello-operator/pkg/apis/github/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestHello(t *testing.T) {
	helloList := &hellov1alpha1.HelloList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Hello",
			APIVersion: "github.awgreene.com/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(hellov1alpha1.AddToScheme, helloList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("hello", func(t *testing.T) {
		t.Run("Cluster", HelloCluster)
		t.Run("Cluster2", HelloCluster)
	})
}

func helloScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	// create hello custom resource
	exampleHello := &hellov1alpha1.Hello{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Hello",
			APIVersion: "github.awgreene.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-hello",
			Namespace: namespace,
		},
		Spec: hellov1alpha1.HelloSpec{
			Size: 3,
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleHello, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for example-hello to reach 3 replicas
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-hello", Namespace: namespace}, exampleHello)
	if err != nil {
		return err
	}

	exampleHello.Spec.Size = 4
	err = f.Client.Update(goctx.TODO(), exampleHello)
	if err != nil {
		return err
	}

	// wait for example-hello to reach 4 replicas
	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-hello", 4, retryInterval, timeout)
}

func HelloCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for hello-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "hello-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = helloScaleTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}
