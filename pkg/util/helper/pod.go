package helper

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	configv1alpha1 "github.com/karmada-io/karmada/pkg/apis/config/v1alpha1"
	"github.com/karmada-io/karmada/pkg/util/lifted"
)

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return GetPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetPodConditionFromList(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

// GeneratePodFromTemplateAndNamespace generate a simple pod object from the given podTemplate and namespace, then
// returns the generated pod.
func GeneratePodFromTemplateAndNamespace(template *corev1.PodTemplateSpec, namespace string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
		},
	}

	pod.Spec = *template.Spec.DeepCopy()
	return pod
}

// GetDependenciesFromPodTemplate extracts the dependencies from the given pod and returns that.
// returns DependentObjectReferences according to the pod, including ConfigMap, Secret, ServiceAccount and PersistentVolumeClaim.
func GetDependenciesFromPodTemplate(podObj *corev1.Pod) ([]configv1alpha1.DependentObjectReference, error) {
	dependentConfigMaps := getConfigMapNames(podObj)
	dependentSecrets := getSecretNames(podObj)
	dependentSas := getServiceAccountNames(podObj)
	dependentPVCs := getPVCNames(podObj)
	var dependentObjectRefs []configv1alpha1.DependentObjectReference
	for cm := range dependentConfigMaps {
		dependentObjectRefs = append(dependentObjectRefs, configv1alpha1.DependentObjectReference{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Namespace:  podObj.Namespace,
			Name:       cm,
		})
	}

	for secret := range dependentSecrets {
		dependentObjectRefs = append(dependentObjectRefs, configv1alpha1.DependentObjectReference{
			APIVersion: "v1",
			Kind:       "Secret",
			Namespace:  podObj.Namespace,
			Name:       secret,
		})
	}

	for sa := range dependentSas {
		dependentObjectRefs = append(dependentObjectRefs, configv1alpha1.DependentObjectReference{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
			Namespace:  podObj.Namespace,
			Name:       sa,
		})
	}

	for pvc := range dependentPVCs {
		dependentObjectRefs = append(dependentObjectRefs, configv1alpha1.DependentObjectReference{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
			Namespace:  podObj.Namespace,
			Name:       pvc,
		})
	}

	return dependentObjectRefs, nil
}

func getSecretNames(pod *corev1.Pod) sets.Set[string] {
	result := sets.New[string]()
	lifted.VisitPodSecretNames(pod, func(name string) bool {
		result.Insert(name)
		return true
	})
	return result
}

func getServiceAccountNames(pod *corev1.Pod) sets.Set[string] {
	result := sets.New[string]()
	if pod.Spec.ServiceAccountName != "" && pod.Spec.ServiceAccountName != "default" {
		result.Insert(pod.Spec.ServiceAccountName)
	}
	return result
}

func getConfigMapNames(pod *corev1.Pod) sets.Set[string] {
	result := sets.New[string]()
	lifted.VisitPodConfigmapNames(pod, func(name string) bool {
		result.Insert(name)
		return true
	})
	return result
}

func getPVCNames(pod *corev1.Pod) sets.Set[string] {
	result := sets.New[string]()
	for i := range pod.Spec.Volumes {
		volume := pod.Spec.Volumes[i]
		if volume.PersistentVolumeClaim != nil {
			claimName := volume.PersistentVolumeClaim.ClaimName
			if len(claimName) != 0 {
				result.Insert(claimName)
			}
		}
	}
	return result
}
