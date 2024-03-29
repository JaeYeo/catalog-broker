/*
Copyright 2019 The Kubernetes Authors.

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

package migration

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sClientSet "k8s.io/client-go/kubernetes"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

// Service provides methods (Backup and Restore) to perform a migration from API Server version (0.2.x) to CRDs version (0.3.0).
type Service struct {
	storagePath        string
	releaseNamespace   string
	apiserverName      string
	webhookServiceName string
	webhookServicePort string

	admInterface  admissionregistrationv1.AdmissionregistrationV1Interface
	appInterface  appsv1.AppsV1Interface
	coreInterface corev1.CoreV1Interface
	scInterface   v1beta1.ServicecatalogV1beta1Interface

	marshaller   func(interface{}) ([]byte, error)
	unmarshaller func([]byte, interface{}) error
}

// NewMigrationService creates a new instance of a Service
func NewMigrationService(scInterface v1beta1.ServicecatalogV1beta1Interface, storagePath string, releaseNamespace string, apiserverName string, webhookServiceName string, webhookServerPort string, k8sclient *k8sClientSet.Clientset) *Service {
	return &Service{
		storagePath:        storagePath,
		releaseNamespace:   releaseNamespace,
		apiserverName:      apiserverName,
		webhookServiceName: webhookServiceName,
		webhookServicePort: webhookServerPort,

		admInterface:  k8sclient.AdmissionregistrationV1(),
		appInterface:  k8sclient.AppsV1(),
		coreInterface: k8sclient.CoreV1(),
		scInterface:   scInterface,

		marshaller: yaml.Marshal,
		unmarshaller: func(b []byte, obj interface{}) error {
			return yaml.Unmarshal(b, obj)
		},
	}
}

// ServiceCatalogResources aggregates all Service Catalog resources
type ServiceCatalogResources struct {
	clusterServiceBrokers []sc.ClusterServiceBroker
	serviceBrokers        []sc.ServiceBroker
	serviceInstances      []sc.ServiceInstance
	serviceBindings       []sc.ServiceBinding
	serviceClasses        []sc.ServiceClass
	servicePlans          []sc.ServicePlan
	clusterServiceClasses []sc.ClusterServiceClass
	clusterServicePlans   []sc.ClusterServicePlan
}

const (
	serviceBrokerFilePrefix        = "servicebroker"
	clusterServiceBrokerFilePrefix = "clusterservicebroker"
	serviceInstanceFilePrefix      = "serviceinstance"
	serviceBindingFilePrefix       = "servicebinding"

	serviceClassFilePrefix        = "serviceclass"
	servicePlanFilePrefix         = "serviceplan"
	clusterServiceClassFilePrefix = "clusterserviceclass"
	clusterServicePlanFilePrefix  = "clusterserviceplan"
)

var (
	// bindingControllerKind contains the schema.GroupVersionKind for this controller type.
	bindingControllerKind = sc.SchemeGroupVersion.WithKind("ServiceBinding")

	finalizersPath = []byte(`[{"op": "remove", "path": "/metadata/finalizers"}]`)

	propagationpolicy = metav1.DeletePropagationOrphan
)

func (r *ServiceCatalogResources) writeMetadata(b *strings.Builder, m metav1.ObjectMeta) {
	b.WriteString("\n\t")
	b.WriteString(m.Namespace)
	b.WriteString("/")
	b.WriteString(m.Name)
}

func (m *Service) loadResource(filename string, obj interface{}) error {
	b, err := os.ReadFile(fmt.Sprintf("%s/%s", m.storagePath, filename))
	if err != nil {
		return fmt.Errorf("while reading file %s/%s: %w", m.storagePath, filename, err)
	}
	err = m.unmarshaller(b, obj)
	if err != nil {
		return fmt.Errorf("while unmarshalling file %s/%s: %w", m.storagePath, filename, err)
	}
	return nil
}

func (m *Service) adjustOwnerReference(om *metav1.ObjectMeta, uidMap map[string]types.UID) {
	if len(om.OwnerReferences) > 0 {
		om.OwnerReferences[0].UID = uidMap[om.OwnerReferences[0].Name]
	}
}

// AssertWebhookServerIsUp make sure webhook server response for request with code 200
func (m *Service) AssertWebhookServerIsUp() error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 3 * time.Second}

	protocol := "http"
	if m.webhookServicePort == "443" {
		protocol = "https"
	}

	return wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, false, func(context.Context) (done bool, err error) {
		url := fmt.Sprintf("%s://%s.%s.svc:%s/mutating-clusterserviceclasses", protocol, m.webhookServiceName, m.releaseNamespace, m.webhookServicePort)
		response, err := client.Get(url)
		if err != nil {
			klog.Infof("while send request to webhook service: %s. Retry...", err)
			return false, nil
		}
		if response.StatusCode != http.StatusOK {
			klog.Infof("Webhook server is not ready. Status code: %d. Retry...", response.StatusCode)
			return false, nil
		}

		klog.Info("Webhook server is ready")
		return true, nil
	})
}

// IsMigrationRequired checks if current version of Service Catalog needs to be migrated
func (m *Service) IsMigrationRequired() (bool, error) {
	_, err := m.appInterface.Deployments(m.releaseNamespace).Get(context.Background(), m.apiserverName, metav1.GetOptions{})
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return false, nil
	default:
		return false, fmt.Errorf("other type of error: %s", err)
	}
	return true, nil
}

// Restore restores Service Catalog resources and adds necessary owner reference to all secrets pointed by service bindings.
func (m *Service) Restore(res *ServiceCatalogResources) error {
	klog.Infof("Applying %d service brokers", len(res.serviceBrokers))

	for _, sb := range res.serviceBrokers {
		err := RetryOnError(retry.DefaultRetry, func() error {
			sb.RecalculatePrinterColumnStatusFields()
			sb.ResourceVersion = ""
			_, err := m.createServiceBroker(sb)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", sb.Name, err)
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	csbNameToUIDMap := map[string]types.UID{}
	klog.Infof("Applying %d cluster service brokers", len(res.clusterServiceBrokers))
	for _, csb := range res.clusterServiceBrokers {
		err := RetryOnError(retry.DefaultRetry, func() error {
			csb.RecalculatePrinterColumnStatusFields()
			csb.ResourceVersion = ""
			created, err := m.createClusterServiceBroker(csb)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", csb.Name, err)
			}
			csbNameToUIDMap[csb.Name] = created.UID

			return nil
		})
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d service classes", len(res.serviceClasses))
	for _, sc := range res.serviceClasses {
		err := RetryOnError(retry.DefaultRetry, func() error {
			sc.ResourceVersion = ""
			sc.UID = ""
			_, err := m.createServiceClass(sc)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", sc.Name, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d cluster service classes", len(res.clusterServiceClasses))
	for _, csc := range res.clusterServiceClasses {
		err := RetryOnError(retry.DefaultRetry, func() error {
			csc.ResourceVersion = ""
			csc.UID = ""
			m.adjustOwnerReference(&csc.ObjectMeta, csbNameToUIDMap)
			_, err := m.createClusterServiceClass(csc)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", csc.Name, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d service plans", len(res.servicePlans))
	for _, sp := range res.servicePlans {
		err := RetryOnError(retry.DefaultRetry, func() error {
			sp.ResourceVersion = ""
			sp.UID = ""
			_, err := m.createServicePlan(sp)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", sp.Name, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d cluster service plans", len(res.clusterServicePlans))
	for _, csp := range res.clusterServicePlans {
		err := RetryOnError(retry.DefaultRetry, func() error {
			csp.ResourceVersion = ""
			csp.UID = ""
			m.adjustOwnerReference(&csp.ObjectMeta, csbNameToUIDMap)
			_, err := m.createClusterServicePlan(csp)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", csp.Name, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d service instances", len(res.serviceInstances))
	for _, si := range res.serviceInstances {
		instance := si.DeepCopy()
		err := RetryOnError(retry.DefaultRetry, func() error {
			si.RecalculatePrinterColumnStatusFields()
			si.ResourceVersion = ""

			// ServiceInstance must not have class/plan refs when it is created
			// These fields must be filled using an update
			si.Spec.ClusterServiceClassRef = nil
			si.Spec.ClusterServicePlanRef = nil
			si.Spec.ServiceClassRef = nil
			si.Spec.ServicePlanRef = nil
			created, err := m.createServiceInstance(si)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", si.Name, err)
			}

			created.Spec.ClusterServiceClassRef = instance.Spec.ClusterServiceClassRef
			created.Spec.ClusterServicePlanRef = instance.Spec.ClusterServicePlanRef
			created.Spec.ServiceClassRef = instance.Spec.ServiceClassRef
			created.Spec.ServicePlanRef = instance.Spec.ServicePlanRef

			updated, err := m.scInterface.ServiceInstances(si.Namespace).Update(context.Background(), created, metav1.UpdateOptions{})
			if err != nil {
				return err
			}

			updated.Status = si.Status
			updated.Status.ObservedGeneration = updated.Generation
			updated, err = m.scInterface.ServiceInstances(si.Namespace).UpdateStatus(context.Background(), updated, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d service bindings", len(res.serviceBindings))
	err := m.RemoveOwnerReferenceFromSecrets()
	if err != nil {
		return fmt.Errorf("when removing owner references from secrets: %w", err)
	}

	for _, sb := range res.serviceBindings {
		err := RetryOnError(retry.DefaultRetry, func() error {
			sb.RecalculatePrinterColumnStatusFields()
			sb.ResourceVersion = ""

			created, err := m.createServiceBinding(sb)
			if err != nil {
				return fmt.Errorf("while restoring %s: %w", sb.Name, err)
			}

			err = m.AddOwnerReferenceToSecret(created)
			if err != nil {
				return fmt.Errorf("when adding owner reference to secret: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadResources loads Service Catalog resources from files.
func (m *Service) LoadResources() (*ServiceCatalogResources, error) {
	files, err := os.ReadDir(m.storagePath)
	if err != nil {
		return nil, err
	}

	var serviceBrokers []sc.ServiceBroker
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceBrokerFilePrefix) {
			var obj sc.ServiceBroker
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceBrokers = append(serviceBrokers, obj)
		}
	}

	var clusterServiceBrokers []sc.ClusterServiceBroker
	for _, file := range files {
		if strings.HasPrefix(file.Name(), clusterServiceBrokerFilePrefix) {
			var obj sc.ClusterServiceBroker
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			clusterServiceBrokers = append(clusterServiceBrokers, obj)
		}
	}

	var serviceInstances []sc.ServiceInstance
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceInstanceFilePrefix) {
			var obj sc.ServiceInstance
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceInstances = append(serviceInstances, obj)
		}
	}

	var serviceBinding []sc.ServiceBinding
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceBindingFilePrefix) {
			var obj sc.ServiceBinding
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceBinding = append(serviceBinding, obj)
		}
	}

	var serviceClasses []sc.ServiceClass
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceClassFilePrefix) {
			var obj sc.ServiceClass
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceClasses = append(serviceClasses, obj)
		}
	}

	var servicePlans []sc.ServicePlan
	for _, file := range files {
		if strings.HasPrefix(file.Name(), servicePlanFilePrefix) {
			var obj sc.ServicePlan
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			servicePlans = append(servicePlans, obj)
		}
	}

	var clusterServiceClasses []sc.ClusterServiceClass
	for _, file := range files {
		if strings.HasPrefix(file.Name(), clusterServiceClassFilePrefix) {
			var obj sc.ClusterServiceClass
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			clusterServiceClasses = append(clusterServiceClasses, obj)
		}
	}

	var clusterServicePlans []sc.ClusterServicePlan
	for _, file := range files {
		if strings.HasPrefix(file.Name(), clusterServicePlanFilePrefix) {
			var obj sc.ClusterServicePlan
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			clusterServicePlans = append(clusterServicePlans, obj)
		}
	}

	return &ServiceCatalogResources{
		serviceBrokers:        serviceBrokers,
		serviceInstances:      serviceInstances,
		serviceBindings:       serviceBinding,
		clusterServiceBrokers: clusterServiceBrokers,
		serviceClasses:        serviceClasses,
		servicePlans:          servicePlans,
		clusterServiceClasses: clusterServiceClasses,
		clusterServicePlans:   clusterServicePlans,
	}, nil
}

// Cleanup deletes all given resources
func (m *Service) Cleanup(resources *ServiceCatalogResources) error {
	klog.Infoln("Cleaning up Service Catalog Resources")
	for _, obj := range resources.serviceBindings {
		err := m.scInterface.ServiceBindings(obj.Namespace).Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ServiceBinding - %s: %w", obj.Name, err)
		}
	}
	for _, obj := range resources.serviceInstances {
		err := m.scInterface.ServiceInstances(obj.Namespace).Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ServiceInstance - %s: %w", obj.Name, err)
		}
	}
	for _, obj := range resources.serviceClasses {
		err := m.scInterface.ServiceClasses(obj.Namespace).Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ServiceClass - %s: %w", obj.Name, err)
		}
	}
	for _, obj := range resources.clusterServiceClasses {
		err := m.scInterface.ClusterServiceClasses().Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ClusterServiceClass - %s: %w", obj.Name, err)
		}
	}
	for _, obj := range resources.servicePlans {
		err := m.scInterface.ServicePlans(obj.Namespace).Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ServicePlan - %s: %w", obj.Name, err)
		}
	}
	for _, obj := range resources.clusterServicePlans {
		err := m.scInterface.ClusterServicePlans().Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ClusterServicePlan - %s: %w", obj.Name, err)
		}
	}
	for _, obj := range resources.serviceBrokers {
		err := m.scInterface.ServiceBrokers(obj.Namespace).Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ServiceBroker - %s: %w", obj.Name, err)
		}
	}
	for _, obj := range resources.clusterServiceBrokers {
		err := m.scInterface.ClusterServiceBrokers().Delete(context.Background(), obj.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("while deleting ClusterServiceBroker - %s: %w", obj.Name, err)
		}
	}
	klog.Infoln("...done")
	return nil
}

func (m *Service) backupResource(obj interface{}, filePrefix string, uid types.UID) error {
	const perm = 0644
	b, err := m.marshaller(obj)
	if err != nil {
		return fmt.Errorf("while marshalling file %s/%s-%s: %w", m.storagePath, filePrefix, uid, err)
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s-%s", m.storagePath, filePrefix, uid), b, perm)
	if err != nil {
		return fmt.Errorf("while writing file %s/%s-%s: %w", m.storagePath, filePrefix, uid, err)
	}
	return nil
}

// BackupResources saves all Service Catalog resources to files.
func (m *Service) BackupResources() (*ServiceCatalogResources, error) {
	klog.Infoln("Saving resources")
	serviceBrokers, err := m.scInterface.ServiceBrokers(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing ServiceBrokers: %w", err)
	}
	for _, sb := range serviceBrokers.Items {
		err := m.backupResource(&sb, serviceBrokerFilePrefix, sb.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ServiceBroker - %s: %w", sb.UID, err)
		}
	}

	clusterServiceBrokers, err := m.scInterface.ClusterServiceBrokers().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing ClusterServiceBrokers: %w", err)
	}
	for _, csb := range clusterServiceBrokers.Items {
		err := m.backupResource(&csb, clusterServiceBrokerFilePrefix, csb.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ClusterServiceBroker - %s: %w", csb.UID, err)
		}
	}

	serviceClasses, err := m.scInterface.ServiceClasses(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing Service Classes: %w", err)
	}
	for _, sc := range serviceClasses.Items {
		err := m.backupResource(&sc, serviceClassFilePrefix, sc.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ServiceClass - %s: %w", sc.UID, err)
		}
	}

	clusterServiceClasses, err := m.scInterface.ClusterServiceClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing ClusterServiceClasses: %w", err)
	}
	for _, csc := range clusterServiceClasses.Items {
		err := m.backupResource(&csc, clusterServiceClassFilePrefix, csc.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ClusterServiceClass - %s: %w", csc.UID, err)
		}
	}

	servicePlans, err := m.scInterface.ServicePlans(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing ServicePlans: %w", err)
	}
	for _, sp := range servicePlans.Items {
		err := m.backupResource(&sp, servicePlanFilePrefix, sp.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ServicePlan - %s: %w", sp.UID, err)
		}
	}

	clusterServicePlans, err := m.scInterface.ClusterServicePlans().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing ClusterServicePlans: %w", err)
	}
	for _, csp := range clusterServicePlans.Items {
		err := m.backupResource(&csp, clusterServicePlanFilePrefix, csp.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ClusterServicePlan - %s: %w", csp.UID, err)
		}
	}

	serviceInstances, err := m.scInterface.ServiceInstances(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing ServiceInstances: %w", err)
	}
	for _, si := range serviceInstances.Items {
		err := m.backupResource(&si, serviceInstanceFilePrefix, si.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ServiceInstance - %s: %w", si.UID, err)
		}
	}

	serviceBindings, err := m.scInterface.ServiceBindings(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("while listing ServiceBindings: %w", err)

	}
	for _, sb := range serviceBindings.Items {
		err := m.backupResource(&sb, serviceBindingFilePrefix, sb.UID)
		if err != nil {
			return nil, fmt.Errorf("while backing up ServiceBinding - %s: %w", sb.UID, err)
		}
	}

	klog.Infoln("...done")
	return &ServiceCatalogResources{
		clusterServiceBrokers: clusterServiceBrokers.Items,
		serviceBrokers:        serviceBrokers.Items,
		clusterServiceClasses: clusterServiceClasses.Items,
		serviceClasses:        serviceClasses.Items,
		clusterServicePlans:   clusterServicePlans.Items,
		servicePlans:          servicePlans.Items,
		serviceInstances:      serviceInstances.Items,
		serviceBindings:       serviceBindings.Items,
	}, nil
}

// AddOwnerReferenceToSecret updates a secret (referenced in the given ServiceBinding) by adding proper owner reference
func (m *Service) AddOwnerReferenceToSecret(sb *sc.ServiceBinding) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		secret, err := m.coreInterface.Secrets(sb.Namespace).Get(context.Background(), sb.Spec.SecretName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		secret.OwnerReferences = []metav1.OwnerReference{
			*metav1.NewControllerRef(sb, bindingControllerKind),
		}
		_, err = m.coreInterface.Secrets(sb.Namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// RemoveOwnerReferenceFromSecrets removes owner references from secrets created for service bindings.
func (m *Service) RemoveOwnerReferenceFromSecrets() error {
	klog.Info("Removing owner references from secrets")
	serviceBindings, err := m.scInterface.ServiceBindings(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sb := range serviceBindings.Items {
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			secret, err := m.coreInterface.Secrets(sb.Namespace).Get(context.Background(), sb.Spec.SecretName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			secret.OwnerReferences = []metav1.OwnerReference{}
			_, err = m.coreInterface.Secrets(sb.Namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
			return err
		})
		if err != nil {
			return err
		}
	}
	klog.Infoln("...done")
	return nil
}

// RetryOnError allows the caller to retry fn in case the error
// according to the provided function. backoff defines the maximum retries and the wait
// interval between two retries.
func RetryOnError(backoff wait.Backoff, fn func() error) error {
	var result error
	err := ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		default:
			result = multierror.Append(result, err)
			return false, nil
		}
	})
	if err == wait.ErrWaitTimeout {
		err = result
	}
	return err
}

// ExponentialBackoff was copied from wait.ExponentialBackoff. Added log messages
func ExponentialBackoff(backoff wait.Backoff, condition wait.ConditionFunc) error {
	i := 0
	for backoff.Steps > 0 {
		i++
		if i > 1 {
			klog.Infof("Retry %d", i)
		}

		if ok, err := condition(); err != nil || ok {
			return err
		}
		if backoff.Steps == 1 {
			break
		}
		time.Sleep(backoff.Step())
	}
	return wait.ErrWaitTimeout
}

func (m *Service) createServiceBroker(cr sc.ServiceBroker) (*sc.ServiceBroker, error) {
	klog.Infof("Processing Service Broker: %s", cr.Name)
	created, err := m.scInterface.ServiceBrokers(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		_, err = m.scInterface.ServiceBrokers(cr.Namespace).Patch(context.Background(), cr.Name, types.JSONPatchType, finalizersPath, metav1.PatchOptions{}, "")
		if err != nil {
			return nil, fmt.Errorf("while removing finalizers from resource '%s': %w", cr.Name, err)
		}

		err = m.scInterface.ServiceBrokers(cr.Namespace).Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ServiceBrokers(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating service broker '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	created.Status = cr.Status
	_, err = m.scInterface.ServiceBrokers(cr.Namespace).UpdateStatus(context.Background(), created, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("while updating status of resource '%s': %w", cr.Name, err)
	}

	return created, nil
}

func (m *Service) createClusterServiceBroker(cr sc.ClusterServiceBroker) (*sc.ClusterServiceBroker, error) {
	klog.Infof("Processing Cluster Service Broker: %s", cr.Name)
	created, err := m.scInterface.ClusterServiceBrokers().Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		_, err = m.scInterface.ClusterServiceBrokers().Patch(context.Background(), cr.Name, types.JSONPatchType, finalizersPath, metav1.PatchOptions{}, "")

		if err != nil {
			return nil, fmt.Errorf("while removing finalizers from resource '%s': %w", cr.Name, err)
		}

		err = m.scInterface.ClusterServiceBrokers().Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ClusterServiceBrokers().Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating resource '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	created.Status = cr.Status
	_, err = m.scInterface.ClusterServiceBrokers().UpdateStatus(context.Background(), created, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("while updating status of cluster service broker '%s': %w", cr.Name, err)
	}

	return created, nil
}

func (m *Service) createServiceClass(cr sc.ServiceClass) (*sc.ServiceClass, error) {
	klog.Infof("Processing Service Class: %s", cr.Name)
	created, err := m.scInterface.ServiceClasses(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		err = m.scInterface.ServiceClasses(cr.Namespace).Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ServiceClasses(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating service broker '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	created.Status = cr.Status
	_, err = m.scInterface.ServiceClasses(cr.Namespace).UpdateStatus(context.Background(), created, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("while updating status of resource '%s': %w", cr.Name, err)
	}

	return created, nil
}

func (m *Service) createClusterServiceClass(cr sc.ClusterServiceClass) (*sc.ClusterServiceClass, error) {
	klog.Infof("Processing Cluster Service Class: %s", cr.Name)
	created, err := m.scInterface.ClusterServiceClasses().Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		err = m.scInterface.ClusterServiceClasses().Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ClusterServiceClasses().Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating resource '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	created.Status = cr.Status
	_, err = m.scInterface.ClusterServiceClasses().UpdateStatus(context.Background(), created, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("while updating status of cluster service broker '%s': %w", cr.Name, err)
	}

	return created, nil
}

func (m *Service) createServicePlan(cr sc.ServicePlan) (*sc.ServicePlan, error) {
	klog.Infof("Processing Service Plan: %s", cr.Name)
	created, err := m.scInterface.ServicePlans(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		err = m.scInterface.ServicePlans(cr.Namespace).Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ServicePlans(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating service broker '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	created.Status = cr.Status
	_, err = m.scInterface.ServicePlans(cr.Namespace).UpdateStatus(context.Background(), created, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("while updating status of resource '%s': %w", cr.Name, err)
	}

	return created, nil
}

func (m *Service) createClusterServicePlan(cr sc.ClusterServicePlan) (*sc.ClusterServicePlan, error) {
	klog.Infof("Processing Cluster Service Plan: %s", cr.Name)
	created, err := m.scInterface.ClusterServicePlans().Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		err = m.scInterface.ClusterServicePlans().Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ClusterServicePlans().Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating resource '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	created.Status = cr.Status
	_, err = m.scInterface.ClusterServicePlans().UpdateStatus(context.Background(), created, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("while updating status of cluster service broker '%s': %w", cr.Name, err)
	}

	return created, nil
}

func (m *Service) createServiceInstance(cr sc.ServiceInstance) (*sc.ServiceInstance, error) {
	klog.Infof("Processing Service Instance: %s", cr.Name)
	created, err := m.scInterface.ServiceInstances(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		_, err = m.scInterface.ServiceInstances(cr.Namespace).Patch(context.Background(), cr.Name, types.JSONPatchType, finalizersPath, metav1.PatchOptions{}, "")
		if err != nil {
			return nil, fmt.Errorf("while removing finalizers from resource '%s': %w", cr.Name, err)
		}

		err = m.scInterface.ServiceInstances(cr.Namespace).Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ServiceInstances(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating service broker '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	return created, nil
}

func (m *Service) createServiceBinding(cr sc.ServiceBinding) (*sc.ServiceBinding, error) {
	klog.Infof("Processing Service Binding: %s", cr.Name)
	created, err := m.scInterface.ServiceBindings(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})

	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		klog.Infof("Resource already exists, deleting and recreating")

		_, err = m.scInterface.ServiceBindings(cr.Namespace).Patch(context.Background(), cr.Name, types.JSONPatchType, finalizersPath, metav1.PatchOptions{}, "")
		if err != nil {
			return nil, fmt.Errorf("while removing finalizers from resource '%s': %w", cr.Name, err)
		}

		err = m.scInterface.ServiceBindings(cr.Namespace).Delete(context.Background(), cr.Name, metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			return nil, fmt.Errorf("while deleting resource '%s': %w", cr.Name, err)
		}

		created, err = m.scInterface.ServiceBindings(cr.Namespace).Create(context.Background(), &cr, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("while re-creating service broker '%s': %w", cr.Name, err)
		}
	default:
		return nil, fmt.Errorf("while creating resource '%s': %w", cr.Name, err)
	}

	created.Status = cr.Status
	_, err = m.scInterface.ServiceBindings(cr.Namespace).UpdateStatus(context.Background(), created, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("while updating status of resource '%s': %w", cr.Name, err)
	}

	return created, nil
}
