/*
Copyright 2025.

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
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcpv1alpha1 "github.com/carlos-gn/smooth-operator/api/v1alpha1"
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Describe("MCPServer Controller", func() {
	Context("When creating an MCPServer", func() {
		It("should create a Deployment with correct spec", func() {
			ctx := context.Background()
			resourceName := fmt.Sprintf("test-deploy-%d", time.Now().UnixNano())

			mcpServer := &mcpv1alpha1.MCPServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPServerSpec{
					Image:    "test-image:v1",
					Replicas: 2,
					Port:     8080,
				},
			}
			Expect(k8sClient.Create(ctx, mcpServer)).Should(Succeed())

			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, deployment)
			}, timeout, interval).Should(Succeed())

			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal("test-image:v1"))
			Expect(*deployment.Spec.Replicas).To(Equal(int32(2)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(int32(8080)))
		})

		It("should create a Service with correct spec", func() {
			ctx := context.Background()
			resourceName := fmt.Sprintf("test-service-%d", time.Now().UnixNano())

			mcpServer := &mcpv1alpha1.MCPServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPServerSpec{
					Image:    "test-image:v1",
					Replicas: 1,
					Port:     8080,
				},
			}
			Expect(k8sClient.Create(ctx, mcpServer)).Should(Succeed())

			service := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, service)
			}, timeout, interval).Should(Succeed())

			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(8080)))
			Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Spec.Selector).To(HaveKeyWithValue("app", resourceName))
		})

		It("should update status phase", func() {
			ctx := context.Background()
			resourceName := fmt.Sprintf("test-status-%d", time.Now().UnixNano())

			mcpServer := &mcpv1alpha1.MCPServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPServerSpec{
					Image:    "test-image:v1",
					Replicas: 1,
					Port:     8080,
				},
			}
			Expect(k8sClient.Create(ctx, mcpServer)).Should(Succeed())

			Eventually(func() string {
				updated := &mcpv1alpha1.MCPServer{}
				_ = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, updated)
				return updated.Status.Phase
			}, timeout, interval).Should(Equal("Pending"))
		})

		It("should have owner references on created resources", func() {
			ctx := context.Background()
			resourceName := fmt.Sprintf("test-owner-%d", time.Now().UnixNano())

			mcpServer := &mcpv1alpha1.MCPServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPServerSpec{
					Image:    "test-image:v1",
					Replicas: 1,
					Port:     8080,
				},
			}
			Expect(k8sClient.Create(ctx, mcpServer)).Should(Succeed())

			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, deployment)
			}, timeout, interval).Should(Succeed())

			Expect(deployment.OwnerReferences).To(HaveLen(1))
			Expect(deployment.OwnerReferences[0].Kind).To(Equal("MCPServer"))
			Expect(deployment.OwnerReferences[0].Name).To(Equal(resourceName))

			service := &corev1.Service{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      resourceName,
			}, service)).Should(Succeed())

			Expect(service.OwnerReferences).To(HaveLen(1))
			Expect(service.OwnerReferences[0].Kind).To(Equal("MCPServer"))
		})
	})

	Context("When updating an MCPServer", func() {
		It("should update Deployment when replicas change", func() {
			ctx := context.Background()
			resourceName := fmt.Sprintf("test-replicas-%d", time.Now().UnixNano())

			mcpServer := &mcpv1alpha1.MCPServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPServerSpec{
					Image:    "test-image:v1",
					Replicas: 1,
					Port:     8080,
				},
			}
			Expect(k8sClient.Create(ctx, mcpServer)).Should(Succeed())

			// Wait for initial Deployment
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, deployment)
			}, timeout, interval).Should(Succeed())

			// Update replicas
			updated := &mcpv1alpha1.MCPServer{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      resourceName,
			}, updated)).Should(Succeed())

			updated.Spec.Replicas = 3
			Expect(k8sClient.Update(ctx, updated)).Should(Succeed())

			// Verify Deployment is updated
			Eventually(func() int32 {
				_ = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, deployment)
				return *deployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(3)))
		})

		It("should update Deployment when image changes", func() {
			ctx := context.Background()
			resourceName := fmt.Sprintf("test-image-%d", time.Now().UnixNano())

			mcpServer := &mcpv1alpha1.MCPServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPServerSpec{
					Image:    "test-image:v1",
					Replicas: 1,
					Port:     8080,
				},
			}
			Expect(k8sClient.Create(ctx, mcpServer)).Should(Succeed())

			// Wait for initial Deployment
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, deployment)
			}, timeout, interval).Should(Succeed())

			// Update image
			updated := &mcpv1alpha1.MCPServer{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      resourceName,
			}, updated)).Should(Succeed())

			updated.Spec.Image = "test-image:v2"
			Expect(k8sClient.Update(ctx, updated)).Should(Succeed())

			// Verify Deployment is updated
			Eventually(func() string {
				_ = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      resourceName,
				}, deployment)
				return deployment.Spec.Template.Spec.Containers[0].Image
			}, timeout, interval).Should(Equal("test-image:v2"))
		})
	})
})
