package grafana

import (
	"fmt"
	"os"
	"path/filepath"

	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	v1 "k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	JsonnetExtension  = ".libsonnet"
	JsonnetAnnotation = "jsonnet/library"
)

func reconcileConfigMaps(cr *grafanav1alpha1.Grafana, r *ReconcileGrafana) error {
	if cr.Spec.Jsonnet == nil || cr.Spec.Jsonnet.LibraryLabelSelector == nil {
		return nil
	}

	selector, err := v13.LabelSelectorAsSelector(cr.Spec.Jsonnet.LibraryLabelSelector)
	if err != nil {
		return err
	}

	configMaps := v1.ConfigMapList{}
	opts := &client.ListOptions{
		LabelSelector: selector,
		Namespace:     cr.Namespace,
	}

	err = r.client.List(r.context, &configMaps, opts)
	if err != nil {
		return err
	}

	jsonnetBasePath := r.config.GetConfigString(config.ConfigJsonnetBasePath, config.JsonnetBasePath)

	for _, configMap := range configMaps.Items {
		if configMap.Annotations[JsonnetAnnotation] != "true" {
			continue
		}

		folderPath, err := createFolder(configMap.Name, jsonnetBasePath)
		if err != nil {
			log.Error(err, fmt.Sprintf("error creating jsonnet library directory for %v", configMap.Name))
			continue
		}

		for filename, contents := range configMap.Data {
			filePath := fmt.Sprintf("%v/%v", folderPath, filename)
			err = createFile(filePath, contents)
			if err != nil {
				return err
			}
			log.Info(fmt.Sprintf("imported jsonnet library %v", filePath))
		}
	}
	return nil
}

func createFolder(configMapName, basePath string) (string, error) {
	folderPath := fmt.Sprintf("%v/%v", basePath, configMapName)
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return folderPath, os.Mkdir(folderPath, os.ModePerm)
	}
	return folderPath, err
}

func createFile(filePath, contents string) error {
	err := validateFileExtension(filePath)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	file.WriteString(contents)
	return file.Close()
}

func validateFileExtension(filePath string) error {
	//check for a valid jsonnet extension
	extension := filepath.Ext(filePath)
	if extension == "" || extension != JsonnetExtension {
		return fmt.Errorf("unkown extention, expected %v", JsonnetExtension)
	}
	return nil
}
