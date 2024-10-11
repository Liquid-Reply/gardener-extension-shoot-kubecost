package kubecost

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	yttcmd "carvel.dev/ytt/pkg/cmd/template"
	yttui "carvel.dev/ytt/pkg/cmd/ui"
	yttfiles "carvel.dev/ytt/pkg/files"
	"github.com/andybalholm/brotli"

	api "github.com/liquid-reply/gardener-extension-shoot-kubecost/pkg/apis/config"
)

//go:embed kubecost.yaml
var manifest string

type KubeCostConfig struct {
	ApiKey string `yaml:"api_key"`
}

var grafanaOverlay string = `#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.subset({"metadata": {"name": "external-grafana-config-map"}})
#@overlay/remove
`

var pvcOverlay string = `#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.subset({"kind": "PersistentVolumeClaim"}), expects="1+"
---
metadata:
  #@overlay/match missing_ok=True
  annotations:
    resources.gardener.cloud/ignore: "true"
`

var labelsOverlay string = `#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.all, expects="1+"
---
metadata:
  #@overlay/match missing_ok=True
  labels:
    #@overlay/match missing_ok=True
    app.kubernetes.io/managed-by: gardener-extension-shoot-kubecost
`

func kubeCostTokenOverlay(token string) string {
	return fmt.Sprintf(`#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.subset({"kind": "ConfigMap", "metadata": {"name": "kubecost-cost-analyzer"}})
---
data:
  kubecost-token: %q
`, token)
}

func Render(config *api.Configuration, compress bool) ([]byte, error) {
	opts := yttcmd.NewOptions()
	noopUI := yttui.NewCustomWriterTTY(false, os.Stderr, os.Stderr)

	var files []*yttfiles.File
	files = append(files, templateAsFile("manifest.yaml", manifest))
	files = append(files, templateAsFile("grafana.yaml", grafanaOverlay))
	files = append(files, templateAsFile("pvc.yaml", pvcOverlay))
	files = append(files, templateAsFile("api-key.yaml", kubeCostTokenOverlay(config.ApiToken)))
	files = append(files, templateAsFile("labels.yaml", labelsOverlay))
	inputs := yttcmd.Input{Files: yttfiles.NewSortedFiles(files)}

	output := opts.RunWithFiles(inputs, noopUI)
	if output.Err != nil {
		return nil, output.Err
	}
	manifest, err := output.DocSet.AsBytes()
	if err != nil {
		return nil, err
	}
	if compress {
		var buf bytes.Buffer
		w := brotli.NewWriterV2(&buf, 7)
		if _, err := w.Write(manifest); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		manifest = buf.Bytes()
	}
	return manifest, nil
}

type noopWriter struct{}

func (w noopWriter) Write(data []byte) (int, error) { return len(data), nil }

func templateAsFile(name, tpl string) *yttfiles.File {
	file, err := yttfiles.NewFileFromSource(yttfiles.NewBytesSource(name, []byte(tpl)))
	if err != nil {
		// should not happen
		panic(err)
	}

	return file
}
