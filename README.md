# K8 Pod Resource Linter

## Overview
The `K8 Pod Resource Linter` is a Kubernetes Custom Admission Webhook Controller that allows us to leverage Kubernetes' [Dynamic Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/). As a `Validating` Admission Webhook prior to a Pod's resource creation, Kubernetes will ensure the resource meets all requirements of any admission controllers on the cluster. It does this by sending a JSON [payload](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admissionreview-request-0) to the Admission Web Hook which must return an HTTP 200 Response with a response body containing the required `AdmissionReview` object fields. For the resource creation to continue to occur. If the Admission Webhook fails to validate the resource Kubernetes will eject the resource from the cluster and prevent its creation from occuring.

## Benefits

## How to Contribute

## How to Deploy
