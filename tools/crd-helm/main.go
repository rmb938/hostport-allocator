package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"
)

func main() {
	log := ctrl.Log.WithName("crd-helm")

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	log.Info("Running Kustomize")
	cmd := exec.Command("kustomize", "build", "config/crd")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "error running kustomize")
		os.Exit(1)
	}

	crdDocs := strings.Split(string(out), "---\n")

	log.Info("Loading Kustomize Output into CRD structs")
	var crdList []v1.CustomResourceDefinition

	for _, crdDoc := range crdDocs {
		var crd v1.CustomResourceDefinition
		err := yaml.Unmarshal([]byte(crdDoc), &crd)
		if err != nil {
			log.Error(err, "error unmarshalling CRDs")
			os.Exit(1)
		}

		crdList = append(crdList, crd)
	}

	for _, crd := range crdList {
		log.Info("Built-In templating CRD", "name", crd.Name)
		if crd.Spec.Conversion != nil {
			if crd.Spec.Conversion.Webhook != nil {
				if crd.Spec.Conversion.Webhook.ClientConfig != nil {
					if crd.Spec.Conversion.Webhook.ClientConfig.Service != nil {
						crd.Spec.Conversion.Webhook.ClientConfig.Service.Name = `{{ include "hostport-allocator.fullname" . }}`
						crd.Spec.Conversion.Webhook.ClientConfig.Service.Namespace = "{{ .Release.Namespace }}"
					}
				}
			}
		}

		log.Info("Marshalling CRD", "name", crd.Name)
		crdBytes, err := yaml.Marshal(crd)
		if err != nil {
			log.Error(err, "error marshalling CRD", "crd", crd.Name)
			os.Exit(1)
		}

		log.Info("String templating CRD", "name", crd.Name)
		crdBytes = append([]byte("{{- if .Values.installCrds }}\n"), crdBytes...)
		crdBytes = append(crdBytes, []byte("{{- end }}\n")...)

		crdBytes = []byte(strings.Replace(string(crdBytes), "caBundle: Cg==", "caBundle: {{ .Values.webhook.caBundle }}", 1))
		crdBytes = []byte(strings.Replace(string(crdBytes), "cert-manager.io/inject-ca-from: $(CERTIFICATE_NAMESPACE)/$(CERTIFICATE_NAME)", `{{ if .Values.webhook.certificate.inject }}cert-manager.io/inject-ca-from: {{ .Release.Namespace }}/{{ include "hostport-allocator.certificateName" . }}{{ end }}`, 1))

		log.Info("Writting CRD", "name", crd.Name)
		crdNameParts := strings.Split(crd.Name, ".")
		ioutil.WriteFile(fmt.Sprintf("deploy/charts/hostport-allocator/templates/crds/%s_%s.yaml", strings.Join(crdNameParts[1:], "."), crdNameParts[0]), crdBytes, 0644)
	}
}
