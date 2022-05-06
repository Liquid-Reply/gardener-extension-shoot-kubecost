module github.com/23technologies/gardener-extension-mwe

go 1.16

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/gardener/gardener v1.43.1
	github.com/go-logr/logr v1.2.2
	github.com/spf13/cobra v1.4.0
	golang.org/x/tools v0.1.9
	k8s.io/api v0.23.5
	k8s.io/code-generator v0.23.5
	k8s.io/component-base v0.23.5
	sigs.k8s.io/controller-runtime v0.11.2
)

replace (
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
	k8s.io/component-base => k8s.io/component-base v0.22.2
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
)
