//go:build e2e
// +build e2e

/*
Copyright 2024 The Kubernetes Authors.

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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
	addonsv1alpha1 "sigs.k8s.io/cluster-api-addon-provider-helm/api/v1alpha1"
	"sigs.k8s.io/cluster-api-addon-provider-helm/controllers/helmreleasedrift"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/cluster-api/test/framework"
)

const (
	initialDeploymentReplicas int32 = 1
	updatedDeploymentReplicas int32 = 2
)

func PatchAndWaitForNginxDeployment(ctx context.Context, spec *addonsv1alpha1.HelmChartProxy, bootstrapClusterProxy framework.ClusterProxy, clusterName string, clusterNamespace string) {
	// Get workload Cluster proxy
	By("creating a clusterctl proxy to the workload cluster")
	workloadClusterProxy := bootstrapClusterProxy.GetWorkloadCluster(ctx, clusterNamespace, clusterName)
	Expect(workloadClusterProxy).NotTo(BeNil())
	client := workloadClusterProxy.GetClient()

	deploymentList := &appsv1.DeploymentList{}
	err := client.List(ctx, deploymentList, ctrlclient.InNamespace(spec.Spec.ReleaseNamespace), ctrlclient.MatchingLabels{helmreleasedrift.InstanceLabelKey: spec.Spec.ReleaseName})
	Expect(err).NotTo(HaveOccurred())
	Expect(deploymentList.Items).NotTo(BeEmpty())

	deployment := &deploymentList.Items[0]
	patch := ctrlclient.MergeFrom(deployment.DeepCopy())
	deployment.Spec.Replicas = ptr.To(updatedDeploymentReplicas)
	err = client.Patch(ctx, deployment, patch)
	Expect(err).NotTo(HaveOccurred())

	deploymentName := ctrlclient.ObjectKeyFromObject(deployment)
	// Wait for Helm release Deployment replicas to be returned back
	Eventually(func() error {
		if err := client.Get(ctx, deploymentName, deployment); err != nil {
			return err
		}
		if *deployment.Spec.Replicas != initialDeploymentReplicas && deployment.Status.ReadyReplicas != initialDeploymentReplicas {
			return fmt.Errorf("expected Deployment replicas to be 1, got %d", deployment.Status.ReadyReplicas)
		}

		return nil
	}, e2eConfig.GetIntervals("default", "wait-helmreleaseproxy")...).Should(Succeed())
}
