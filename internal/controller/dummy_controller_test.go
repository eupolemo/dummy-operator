package controller

import (
	"context"
	"fmt"
	dummyv1alpha1 "github.com/eupolemo/dummy-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var _ = Describe("Dummy controller", func() {
	Context("Dummy controller test", func() {

		const DummyName = "test-dummy"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DummyName,
				Namespace: DummyName,
			},
		}

		typeNamespaceName := types.NamespacedName{Name: DummyName, Namespace: DummyName}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Setting the Image ENV VAR which stores the Operand image")
			err = os.Setenv("DUMMY_IMAGE", "docker.io/eupolemo/dummy:latest")
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)

			By("Removing the Image ENV VAR which stores the Operand image")
			_ = os.Unsetenv("DUMMY_IMAGE")
		})

		It("should successfully reconcile a custom resource for Dummy", func() {
			By("Creating the custom resource for the Kind Dummy")
			dummy := &dummyv1alpha1.Dummy{}
			err := k8sClient.Get(ctx, typeNamespaceName, dummy)
			if err != nil && errors.IsNotFound(err) {
				dummy := &dummyv1alpha1.Dummy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      DummyName,
						Namespace: namespace.Name,
					},
					Spec: dummyv1alpha1.DummySpec{
						Message: "I'm just a dummy",
					},
				}

				err = k8sClient.Create(ctx, dummy)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &dummyv1alpha1.Dummy{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			dummyReconciler := &DummyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = dummyReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if Pod was successfully created in the reconciliation")
			Eventually(func() error {
				found := &corev1.Pod{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking the latest Status added to the Dummy instance")
			Eventually(func() error {
				if dummy.Status.PodStatus == "Pending" {
					return fmt.Errorf("The latest podStatus (%s) is not created", dummy.Status.PodStatus)
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
		})
	})
})
