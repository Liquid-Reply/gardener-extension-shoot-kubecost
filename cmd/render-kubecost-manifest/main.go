package main

import (
	"fmt"

	"github.com/liquid-reply/gardener-extension-shoot-kubecost/kubecost"
)

func main() {
	out := kubecost.Render(kubecost.KubeCostConfig{
		ApiKey: "my-kubecost-key",
	})
	fmt.Println(string(out))
}
