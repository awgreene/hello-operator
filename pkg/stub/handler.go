package stub

import (
	"context"
	"fmt"
	"reflect"

	"github.com/awgreene/hello-operator/pkg/apis/github/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func NewHandler(m *Metrics) sdk.Handler {
	return &Handler{
		metrics: m,
	}
}

type Metrics struct {
	operatorErrors prometheus.Counter
}

type Handler struct {
	// Metrics example
	metrics *Metrics
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Hello:
		hello := o

		// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
		// All secondary resources must have the CR set as their OwnerReference for this to be the case
		if event.Deleted {
			return nil
		}
		// Create the deployment if it doesn't exist
		dep := deploymentForHello(hello)
		err := sdk.Create(dep)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create deployment: %v", err)
		}

		// Ensure the deployment size is the same as the spec
		err = sdk.Get(dep)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %v", err)
		}
		size := hello.Spec.Size
		world := hello.Spec.World

		if *dep.Spec.Replicas != size || dep.Spec.Template.Spec.Containers[0].Env[0].Value != world {
			dep.Spec.Replicas = &size
			dep.Spec.Template.Spec.Containers[0].Env[0].Value = world
			err = sdk.Update(dep)
			if err != nil {
				return fmt.Errorf("failed to update deployment: %v", err)
			}
		}

		podList := podList()
		labelSelector := labels.SelectorFromSet(labelsForHello(hello.Name)).String()
		listOps := &metav1.ListOptions{LabelSelector: labelSelector}
		err = sdk.List(hello.Namespace, podList, sdk.WithListOptions(listOps))
		if err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}
		podNames := getPodNames(podList.Items)
		if !reflect.DeepEqual(podNames, hello.Status.Nodes) {
			hello.Status.Nodes = podNames
			err := sdk.Update(hello)
			if err != nil {
				return fmt.Errorf("failed to update memcached status: %v", err)
			}
		}
	}
	return nil
}

// deploymentForHello returns a Hello Deployment object
func deploymentForHello(m *v1alpha1.Hello) *appsv1.Deployment {
	ls := labelsForHello(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Image:   "docker.io/agreene/hello-go", // The image the operator deploys
						Name:    "hello",
						Command: []string{"go", "run", "main.go"},
						Ports: []v1.ContainerPort{{
							ContainerPort: 8000, // The port the container listens on
							Name:          "hello",
						}},
						Env: []v1.EnvVar{{
							Name:  "WORLD", // The WORLD environment variable
							Value: m.Spec.World,
						}},
					}},
				},
			},
		},
	}
	addOwnerRefToObject(dep, asOwner(m))
	return dep
}

// labelsForHello returns the labels for selecting the resources
// belonging to the given hello CR name.
func labelsForHello(name string) map[string]string {
	return map[string]string{"app": "hello", "hello_cr": name}
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// asOwner returns an OwnerReference set as the memcached CR
func asOwner(m *v1alpha1.Hello) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: m.APIVersion,
		Kind:       m.Kind,
		Name:       m.Name,
		UID:        m.UID,
		Controller: &trueVar,
	}
}

// podList returns a v1.PodList object
func podList() *v1.PodList {
	return &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []v1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func RegisterOperatorMetrics() (*Metrics, error) {
	operatorErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "memcached_operator_reconcile_errors_total",
		Help: "Number of errors that occurred while reconciling the memcached deployment",
	})
	err := prometheus.Register(operatorErrors)
	if err != nil {
		return nil, err
	}
	return &Metrics{operatorErrors: operatorErrors}, nil
}
