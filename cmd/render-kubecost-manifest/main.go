package main

import (
	"fmt"

	"github.com/liquid-reply/gardener-extension-shoot-kubecost/kubecost"
)

func main() {
	out, err := kubecost.Render(kubecost.KubeCostConfig{
		ApiKey: "my-kubecost-key",
	}, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
