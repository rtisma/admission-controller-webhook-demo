package main

import (
	"fmt"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	"log"
)

// Checks of a pod spec contains a volume with
func hasVolume(pod *v1.Pod, targetVolumeName string) bool {
	if pod.Spec.Volumes != nil {
		for _, volume := range pod.Spec.Volumes {
			if volume.Name == targetVolumeName {
				return true
			}
		}
	}
	return false
}

func findTargetContainer(pod *v1.Pod, targetContainerName string) (*v1.Container, int, error) {
	if pod.Spec.Containers != nil {
		for pos, container := range pod.Spec.Containers {
			if container.Name == targetContainerName {
				return &container, pos, nil
			}
		}
	}
	return nil, -1, fmt.Errorf("container with name %s does not exist", targetContainerName)
}

func extractPodSpec(req *v1beta1.AdmissionRequest) (v1.Pod, error) {
	pod := v1.Pod{}
	// This handler should only get called on Pod objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource != podResource {
		return pod, fmt.Errorf("expect resource to be %s", podResource)
	}

	// Parse the Pod object.
	raw := req.Object.Raw
	log.Println("ROB_DUMPPPPPPPPPPPPPPPPPP", string(raw))
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return pod, fmt.Errorf("could not deserialize pod object: %v", err)
	}
	return pod, nil
}

func hasVolumeMount(container *v1.Container) bool {
	for _, volMount := range container.VolumeMounts {
		if volMount.Name == scratchVolumeName {
			return true
		}
	}
	return false
}

func findVolumeMount(container *v1.Container) (*v1.VolumeMount, int) {
	for pos, volMount := range container.VolumeMounts {
		if volMount.Name == scratchVolumeName {
			return &volMount, pos
		}
	}
	return nil, -1
}

