package constants

const (
	// ExtensionType is the name of the extension type.
	ExtensionType = "shoot-kubecost"
	// ServiceName is the name of the service.
	ServiceName = ExtensionType

	extensionServiceName = "extension-" + ServiceName

	ManagedResourceNameKubeCostConfig = extensionServiceName + "-shoot-kubecost-config"
)
