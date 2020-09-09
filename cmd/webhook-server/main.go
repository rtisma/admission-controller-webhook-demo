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
	"encoding/json"
	// "errors"
	"fmt"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"strconv"
)

const (
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

// applySecurityDefaults implements the logic of our example admission controller webhook. For every pod that is created
// (outside of Kubernetes namespaces), it first checks if `runAsNonRoot` is set. If it is not, it is set to a default
// value of `false`. Furthermore, if `runAsUser` is not set (and `runAsNonRoot` was not initially set), it defaults
// `runAsUser` to a value of 1234.
//
// To demonstrate how requests can be rejected, this webhook further validates that the `runAsNonRoot` setting does
// not conflict with the `runAsUser` setting - i.e., if the former is set to `true`, the latter must not be `0`.
// Note that we combine both the setting of defaults and the check for potential conflicts in one webhook; ideally,
// the latter would be performed in a validating webhook admission controller.
type EmptyDirData struct {
	Name string `json:"name"`
	EmptyDir interface{} `json:"emptyDir"`
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

	//TODO: rtisma not sure if this is right
	//rtisma   pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
	var emptyDirData = EmptyDirData{ Name: scratchVolumeName}
	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/spec/volumes",
		Value: emptyDirData,
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
	fmt.Println("POD: ",string(b))


	//TODO: iterate and dump patches, so they can be applied to incomming request. render and check its ok
	b2, err2 := json.Marshal(patches)
	if  err2 == nil {
		fmt.Println("Patches: ", string(b2))
	}

	return patches, nil
}

func main() {
	var cfg = parseConfig()

	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(applySecurityDefaults))
	server := &http.Server{
		// We listen on port 8443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":"+cfg.Server.Port,
		Handler: mux,
	}

	log.Println("Starting server on port ",cfg.Server.Port," with TLS ENABLED=",cfg.Server.SSL.Enable)
	if cfg.Server.SSL.Enable {
		log.Fatal(server.ListenAndServeTLS(cfg.Server.SSL.CertPath, cfg.Server.SSL.KeyPath))
	} else {
		log.Fatal(server.ListenAndServe())
	}
	log.Println("Stopped server")
}

