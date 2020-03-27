package sts

import (
	"github.com/mesosphere/kudo-cassandra-operator/images/cassandra-recovery/pkg/client"


	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"log"
)

func Process(item interface{}) {
	if item == nil {
		// Event was deleted
	} else if pod, ok := detectRecoveryConditions(item); ok {
		log.Printf("the pod %s/%s meets the recovery conditions.\n", pod.Namespace, pod.Name)
		cleanStartPod(pod)
	}
}

func detectRecoveryConditions(item interface{}) (*v1.Pod, bool) {
	config, _ := client.GetKubernetesClient()
	clientSet, _ := kubernetes.NewForConfig(config)
	if event, ok := item.(*v1.Event); ok {
		if event.InvolvedObject.Kind == "Pod" && event.Reason == "FailedScheduling" {
			log.Printf("FailedScheduling detected: %s for %s/%s.\n", event.Name, event.InvolvedObject.Namespace, event.InvolvedObject.Name)
			pod, _ := clientSet.CoreV1().Pods(event.InvolvedObject.Namespace).Get(event.InvolvedObject.Name, meta_v1.GetOptions{})
			nodeDown, _ := detectNodeDown(clientSet, pod)
			if nodeDown {
				log.Printf("Node is down for %s/%s.\n", event.InvolvedObject.Namespace, event.InvolvedObject.Name)
				return pod, true
			}
			pvcDown, _ := detectPVCDown(clientSet, pod)
			if pvcDown {
				log.Printf("PVC isn't still created for %s/%s.\n", event.InvolvedObject.Namespace, event.InvolvedObject.Name)
				return pod, true
			}
		}
	}
	return nil, false
}

func detectPVCDown(clientSet *kubernetes.Clientset, pod *v1.Pod) (bool, error) {
	// we need to check if PVC is still deleted
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			_, err := clientSet.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(vol.PersistentVolumeClaim.ClaimName, meta_v1.GetOptions{})
			if errors.IsNotFound(err) {
				return true, nil
			}
			if err != nil {
				return false, err
			}
		}
	}

	return false, nil
}

func detectNodeDown(clientSet *kubernetes.Clientset, pod *v1.Pod) (bool, error) {
	// we cannot check by node name here as the node will be  Nil here
	// we need to check through PVC
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			pvc, err := clientSet.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(vol.PersistentVolumeClaim.ClaimName, meta_v1.GetOptions{})
			if err != nil {
				return false, err
			}
			pv, err := clientSet.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, meta_v1.GetOptions{})
			if err != nil{
				return false, err
			}
			for _,node  := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms{
				for _,expr := range node.MatchExpressions{
					if expr.Key == "kubernetes.io/hostname" {
						log.Printf("found match expression: %+v", expr)
						// TODO improve if there are more nodes
						node, err := clientSet.CoreV1().Nodes().Get(expr.Values[0], meta_v1.GetOptions{})
						if err != nil {
							if errors.IsNotFound(err) {
								return true, nil
							} else {
								return false, err
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

func cleanStartPod(pod *v1.Pod) {
	config, _ := client.GetKubernetesClient()
	clientSet, _ := kubernetes.NewForConfig(config)
	pvcs := getPVCs(pod)
	for _, pvc := range pvcs {
		clientSet.CoreV1().PersistentVolumeClaims(pod.Namespace).Delete(pvc.ClaimName, &meta_v1.DeleteOptions{})
	}
	clientSet.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &meta_v1.DeleteOptions{})
}

func getPVCs(pod *v1.Pod) []*v1.PersistentVolumeClaimVolumeSource {
	pvcs := []*v1.PersistentVolumeClaimVolumeSource{}
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			pvcs = append(pvcs, vol.PersistentVolumeClaim)
			deattachPV(vol.PersistentVolumeClaim, pod.Namespace)
		}
	}
	return pvcs
}

func deattachPV(source *v1.PersistentVolumeClaimVolumeSource, namespace string) {
	config, _ := client.GetKubernetesClient()
	clientSet, _ := kubernetes.NewForConfig(config)
	pvc, _ := clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(source.ClaimName, meta_v1.GetOptions{})

	pv, _ := clientSet.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, meta_v1.GetOptions{})

	pv.Spec.ClaimRef = nil
	_, err := clientSet.CoreV1().PersistentVolumes().Update(pv)
	if err != nil {
		log.Printf("PV claimRef cleared for PV:%s/%s", pv.Namespace, pv.Name)
	}
}
