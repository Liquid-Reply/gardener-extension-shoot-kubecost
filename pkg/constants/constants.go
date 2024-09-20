package constants

const (
	// ExtensionType is the name of the extension type.
	ExtensionType = "shoot-flux"
	// ServiceName is the name of the service.
	ServiceName = ExtensionType

	extensionServiceName = "extension-" + ServiceName

	ManagedResourceNameKubeCostConfig = extensionServiceName + "-shoot-kubecost-config"
)
