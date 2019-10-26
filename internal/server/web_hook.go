package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
)

func (e *Env) handleValidationWebHookRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a handleValidationWebHookRequest")
	var body []byte
	if r.Body != nil {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("ioutil.ReadAll: %v", err)
			http.Error(w, "error reading body", http.StatusBadRequest)
			return
		}
		body = data
	}
	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	_, _, err := deserializer.Decode(body, nil, &ar)
	if err != nil {
		log.Printf("deserializer.Decode ERR: %v", err)
		http.Error(w, "error deserailizing body", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		log.Print("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	var (
		containers []corev1.Container
	)

	log.Printf("Checking request type \n Request: \n %#v \n\n", ar)

	switch ar.Request.Kind.Kind {
	case "Pod":
		log.Printf("Admission Review Request Kind is of type Pod: %v", ar.Request.Kind.Kind)
		var pod corev1.Pod
		err = json.Unmarshal(ar.Request.Object.Raw, &pod)
		if err != nil {
			admissionResponse = &v1beta1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		containers = pod.Spec.Containers
	default:
		log.Printf("Admission Review Request IS NOT OF type Pod: %v", ar.Request.Kind.Kind)
		admissionResponse = &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	if containers == nil {
		log.Printf("container sliced contained no containers! weird")
		admissionResponse = &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Reason: "Pods contained no containers! Nothing to lint resources against.",
			},
		}
	} else {
		log.Printf("Containers is not nil: \n\n %#v \n\n", containers)
		for _, c := range containers {
			log.Printf("Current Container: \n\n %#v \n\n", c)
			cpuReq := c.Resources.Requests.Cpu()
			memReq := c.Resources.Requests.Memory()
			cpuLim := c.Resources.Limits.Cpu()
			memLim := c.Resources.Limits.Memory()
			cpuReq.Value()
			if cpuReq.Value() == 0 || memReq.Value() == 0 || cpuLim.Value() == 0 || memLim.Value() == 0 {
				log.Printf("Validating Admission Web Hook found a pod without resource requests or limits")
				log.Printf("Ejecting container name: %v from the cluster", c.Name)
				rejectionReason := fmt.Sprintf("container: %v lacks the appropriate resource request or limit "+
					"definitions", c.Name)
				admissionResponse = &v1beta1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Reason: metav1.StatusReason(rejectionReason),
					},
				}
				break
			}
			log.Printf("All the Resource fields are defined for! Container: %v", c.Name)
			log.Printf("\n cpuReq: %v \n memReq: %v \n cpuLim: %v \n memLim: %v \n", cpuReq, memReq,
				cpuLim, memLim)
		}

		log.Printf("Finished iteratering through all the pod's contaienrs, returning allowed admission review")
		if admissionResponse == nil {
			log.Print("Admission response is nil, no errors for container found, validating pod admission")
			admissionResponse = &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}

	}

	log.Printf("\n\n Admission Request UID is: %+v \n\n", ar.Request.UID)
	log.Printf("\n\n admissionResponse is: \n\n %+v \n\n", admissionResponse)

	admissionReview := v1beta1.AdmissionReview{}

	admissionReview.Response = admissionResponse
	if ar.Request != nil {
		admissionReview.Response.UID = ar.Request.UID
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Printf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	log.Print("Ready to write reponse ...")
	log.Printf("Response Body Contents: \n\n %v \n\n", string(resp))
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}

}
