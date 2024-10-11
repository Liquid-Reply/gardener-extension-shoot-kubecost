package main

import (
	"fmt"

	"github.com/liquid-reply/gardener-extension-shoot-kubecost/kubecost"
	apisconfig "github.com/liquid-reply/gardener-extension-shoot-kubecost/pkg/apis/config"
)

func main() {
	out, err := kubecost.Render(&apisconfig.Configuration{
		ApiToken: "my-kubecost-key",
	}, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
