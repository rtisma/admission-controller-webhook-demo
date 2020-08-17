/*
Copyright (c) 2019 StackRox Inc.

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

package main

import (
	// "errors"
	"fmt"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

const (
	useTLS = false
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

var (
	podResource    = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
	overrideVolumePathCollision = true
	targetContainerName = "busybox"
	scratchDirName = "/icgc-argo-scratch"
	scratchVolumeName = "icgc-argo-scratch"
)

// Checks of a pod spec contains a volume with
func hasVolume(pod *corev1.Pod, targetVolumeName string) bool {
	if pod.Spec.Volumes != nil {
		for _, volume := range pod.Spec.Volumes {
			if volume.Name == targetVolumeName {
				return true
			}
		}
	}
	return false
}

func findTargetContainer(pod *corev1.Pod, targetContainerName string) (*corev1.Container, int, error) {
	if pod.Spec.Containers != nil {
		for pos, container := range pod.Spec.Containers {
			if container.Name == targetContainerName {
				return &container, pos, nil
			}
		}
	}
	return nil, -1, fmt.Errorf("container with name %s does not exist", targetContainerName)
}

func extractPodSpec(req *v1beta1.AdmissionRequest) (corev1.Pod, error){
	pod := corev1.Pod{}
	// This handler should only get called on Pod objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource != podResource {
		return pod, fmt.Errorf( "expect resource to be %s", podResource)
	}

	// Parse the Pod object.
	raw := req.Object.Raw
	log.Println("ROB_DUMPPPPPPPPPPPPPPPPPP", string(raw))
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return pod, fmt.Errorf("could not deserialize pod object: %v", err)
	}
	return pod, nil
}


// applySecurityDefaults implements the logic of our example admission controller webhook. For every pod that is created
// (outside of Kubernetes namespaces), it first checks if `runAsNonRoot` is set. If it is not, it is set to a default
// value of `false`. Furthermore, if `runAsUser` is not set (and `runAsNonRoot` was not initially set), it defaults
// `runAsUser` to a value of 1234.
//
// To demonstrate how requests can be rejected, this webhook further validates that the `runAsNonRoot` setting does
// not conflict with the `runAsUser` setting - i.e., if the former is set to `true`, the latter must not be `0`.
// Note that we combine both the setting of defaults and the check for potential conflicts in one webhook; ideally,
// the latter would be performed in a validating webhook admission controller.

func hasVolumeMount(container *corev1.Container) bool {
	for _, volMount:= range container.VolumeMounts {
		if volMount.Name == scratchVolumeName{
			return true
		}
	}
	return false
}

func findVolumeMount(container *corev1.Container) (*corev1.VolumeMount, int) {
	for pos, volMount:= range container.VolumeMounts {
		if volMount.Name == scratchVolumeName{
			return &volMount, pos
		}
	}
	return nil, -1
}

func applySecurityDefaults(req *v1beta1.AdmissionRequest) ([]patchOperation, error) {

	//var pod *corev1.Pod
	//var err *error
	var patches []patchOperation
	var pod, err = extractPodSpec(req)
	if err != nil {
		return patches, err
	}

	if hasVolume(&pod, scratchVolumeName){
		log.Println("Already contains the scratch volume name: ", scratchVolumeName)
		return patches, nil
	}

	var volumeSource = corev1.VolumeSource{EmptyDir: nil}
	var volume  = corev1.Volume{Name: scratchVolumeName, VolumeSource: volumeSource}

	//TODO: rtisma not sure if this is right
	//rtisma   pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/spec/volumes",
		Value: volume,
	})

	var container, containerPos, err2 = findTargetContainer(&pod, targetContainerName)
	if err2 != nil {
		log.Println("Did not find container with name '",targetContainerName,"'. Skipping mutation")
		return patches, nil
	}

	var containerVolumeMount, volumeMountPos = findVolumeMount(container)
	var volumeMount = corev1.VolumeMount{Name: scratchVolumeName, MountPath: scratchDirName}
	if containerVolumeMount == nil{
		//rtisma  container.VolumeMounts = append(container.VolumeMounts, volumeMount)
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/containers/"+strconv.Itoa(containerPos)+"/volumeMounts",
			Value: volumeMount,
		})

	} else {
		if overrideVolumePathCollision{
			log.Println("Container volume mount ",scratchVolumeName," already exists but overriding ")
			//rtisma    containerVolumeMount = &volumeMount
			patches = append(patches, patchOperation{
				Op:    "replace",
				Path:  "/spec/containers/"+strconv.Itoa(containerPos)+"/volumeMounts/"+strconv.Itoa(volumeMountPos),
				Value: volumeMount,
			})
		} else {
			log.Println("Container volume mount ",scratchVolumeName," already exists, and NOT overriding ")
		}
	}

	//TEST
	b, err := json.Marshal(pod)
	if err != nil {
		fmt.Println(err)
		return patches, nil
	}
	fmt.Println(string(b))

	return patches, nil




	//// Retrieve the `runAsNonRoot` and `runAsUser` values.
	//var runAsNonRoot *bool
	//var runAsUser *int64
	//
	//
	//}
	//if pod.Spec.SecurityContext != nil {
	//	runAsNonRoot = pod.Spec.SecurityContext.RunAsNonRoot
	//	runAsUser = pod.Spec.SecurityContext.RunAsUser
	//}
	//
	//// Create patch operations to apply sensible defaults, if those options are not set explicitly.
	//if runAsNonRoot == nil {
	//	patches = append(patches, patchOperation{
	//		Op:    "add",
	//		Path:  "/spec/securityContext/runAsNonRoot",
	//		// The value must not be true if runAsUser is set to 0, as otherwise we would create a conflicting
	//		// configuration ourselves.
	//		Value: runAsUser == nil || *runAsUser != 0,
	//	})
	//
	//	if runAsUser == nil {
	//		patches = append(patches, patchOperation{
	//			Op:    "add",
	//			Path:  "/spec/securityContext/runAsUser",
	//			Value: 1234,
	//		})
	//	}
	//} else if *runAsNonRoot == true && (runAsUser != nil && *runAsUser == 0) {
	//	// Make sure that the settings are not contradictory, and fail the object creation if they are.
	//	return nil, errors.New("runAsNonRoot specified, but runAsUser set to 0 (the root user)")
	//}

}

func main() {
	var certPath = ""
	var keyPath = ""
	if useTLS{
		certPath = filepath.Join(tlsDir, tlsCertFile)
		keyPath = filepath.Join(tlsDir, tlsKeyFile)
	}

	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(applySecurityDefaults))
	server := &http.Server{
		// We listen on port 8443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":8443",
		Handler: mux,
	}
	if useTLS{
		log.Println("Starting server on port 8443 with TLS ENABLED")
		log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
	} else {
		log.Println("Starting server on port 8443 with TLS DISABLED")
		log.Fatal(server.ListenAndServe())
	}
	log.Println("Stopped server")
}
