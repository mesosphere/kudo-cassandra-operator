package sts

import (
	"fmt"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var evictionLabel = os.Getenv("EVICTION_LABEL")

func Process(client *kubernetes.Clientset, item interface{}) {
	if item == nil {
		// Event was deleted
		return
	}

	pod, ok := item.(*corev1.Pod)
	if !ok {
		// We only act on Pods
		return
	}

	needsRecovery, err := detectRecoveryConditions(client, pod)
	if err != nil {
		log.Printf("ERROR: failed to detect recovery condition: %v", err)
	}

	if needsRecovery {
		log.Printf("the pod %s/%s meets the recovery conditions.\n", pod.Namespace, pod.Name)
		err = cleanStartPod(client, pod)
	} else if detectEvictionCondition(pod) {
		log.Printf("the pod %s/%s meets the eviction conditions.\n", pod.Namespace, pod.Name)
		err = cleanStartPod(client, pod)
	}
	if err != nil {
		log.Printf("ERROR: Failed to clean start pod: %v", err)
	}
}

func detectEvictionCondition(pod *corev1.Pod) bool {
	if val, ok := pod.Labels[evictionLabel]; ok {
		if val == "true" {
			log.Printf("Pod %s has eviction label set", pod.Name)
			return true
		}
	}
	return false
}

func detectRecoveryConditions(client *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	isUnschedulable := false
	for _, condition := range pod.Status.Conditions {
		if condition.Type == "PodScheduled" && condition.Reason == "Unschedulable" {
			isUnschedulable = true
			break
		}
	}

	if isUnschedulable {
		log.Printf("FailedScheduling detected for %s/%s.\n", pod.Namespace, pod.Name)
		nodeDown, err := detectNodeDown(client, pod)
		if err != nil {
			return false, fmt.Errorf("failed to detect if node for pod %s/%s is down: %v", pod.Namespace, pod.Name, err)
		}
		if nodeDown {
			log.Printf("Node is down for %s/%s.\n", pod.Namespace, pod.Name)
			return true, nil
		}
		pvcDown, err := detectPVCDown(client, pod)
		if err != nil {
			return false, fmt.Errorf("failed to detect if pv for pod %s/%s is down: %v", pod.Namespace, pod.Name, err)
		}
		if pvcDown {
			log.Printf("PVC isn't still created for %s/%s.\n", pod.Namespace, pod.Name)
			return true, nil
		}
	}

	return false, nil
}

func detectPVCDown(client *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	// we need to check if PVC is still deleted
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			_, err := client.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(vol.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				return true, nil
			}
			if err != nil {
				return false, fmt.Errorf("failed to retrieve pvc %s/%s: %v", pod.Namespace, vol.PersistentVolumeClaim.ClaimName, err)
			}
		}
	}

	return false, nil
}

func detectNodeDown(client *kubernetes.Clientset, pod *corev1.Pod) (bool, error) {
	// we cannot check by node name here as the node will be  Nil here
	// we need to check through PVC
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {

			pvc, err := client.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(vol.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("failed to get pvc %s/%s: %v", pod.Namespace, vol.PersistentVolumeClaim.ClaimName, err)
			}

			pv, err := client.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("failed to get PV %s: %v", pvc.Spec.VolumeName, err)
			}

			for _, node := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
				for _, expr := range node.MatchExpressions {
					if expr.Key == "kubernetes.io/hostname" {
						log.Printf("Found required hostname affinity for PV %s: %+v", pv.Name, expr)

						if len(expr.Values) > 1 {
							log.Printf("WARN: Required NodeAffinity for PV %s has more than one value for hostname: %v", pv.Name, expr.Values)
						}

						node, err := client.CoreV1().Nodes().Get(expr.Values[0], metav1.GetOptions{})
						if err != nil {
							if errors.IsNotFound(err) {
								return true, nil
							} else {
								return false, fmt.Errorf("failed to get node %s: %v", expr.Values[0], err)
							}
						}
						if node == nil {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}

func cleanStartPod(client *kubernetes.Clientset, pod *corev1.Pod) error {
	// Get all PVCs from the pod
	pvcs, err := getPVCs(client, pod)
	if err != nil {
		return fmt.Errorf("failed to get PVCs from pod %s/%s: %v", pod.Namespace, pod.Name)
	}

	// Detach PVCs from their PVs
	for _, pvc := range pvcs {
		err := detachPVCFromPV(client, pvc)
		if err != nil {
			return fmt.Errorf("failed to detach PV from PVC %s/%s: %v", pvc.Namespace, pvc.Name, err)
		}
	}

	// Delete PVCs
	for _, pvc := range pvcs {
		err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Delete(pvc.Name, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete PVC %s/%s: %v", pvc.Namespace, pvc.Name, err)
		}
	}

	// Delete pod to allow for rescheduling
	err = client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod %s/%s for rescheduling: %s", pod.Namespace, pod.Name, err)
	}

	return nil
}

func getPVCs(client *kubernetes.Clientset, pod *corev1.Pod) ([]*corev1.PersistentVolumeClaim, error) {
	pvcs := make([]*corev1.PersistentVolumeClaim, len(pod.Spec.Volumes))
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			source := vol.PersistentVolumeClaim
			pvc, err := client.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(source.ClaimName, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				log.Printf("Unable to find PVC %s/%s. Assuming it was deleted in a previous invocation.", pod.Namespace, source.ClaimName)
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("failed to get PVC %s/%s: %v", pod.Namespace, source.ClaimName, err)
			}
			pvcs = append(pvcs, pvc)
		}
	}
	return pvcs, nil
}

func detachPVCFromPV(client *kubernetes.Clientset, pvc *corev1.PersistentVolumeClaim) error {
	pv, err := client.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get PV %s: %v", pvc.Spec.VolumeName, err)
	}

	log.Printf("Deattach PVC %s/%s from PV %s: ", pvc.Namespace, pvc.Name, pv.Name)
	pv.Spec.ClaimRef = nil

	_, err = client.CoreV1().PersistentVolumes().Update(pv)
	if err != nil {
		return fmt.Errorf("failed to clear claimRef for PV %s: %v", pv.Name, err)
	}

	log.Printf("PV claimRef cleared for PV %s", pv.Name)
	return nil
}
