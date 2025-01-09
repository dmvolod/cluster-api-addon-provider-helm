package main

import (
	"context"
	"log"

	addonsv1alpha1 "sigs.k8s.io/cluster-api-addon-provider-helm/api/v1alpha1"
	"sigs.k8s.io/cluster-api-addon-provider-helm/internal"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	cfg := config.GetConfigOrDie()
	println(cfg.Host)

	clnt := internal.HelmClient{}
	hrpSpec := addonsv1alpha1.HelmReleaseProxySpec{
		ChartName:        "nginx-ingress",
		RepoURL:          "https://helm.nginx.com/stable",
		ReleaseName:      "my-release",
		ReleaseNamespace: "default",
		Values:           "",
	}
	rel, err := clnt.TemplateHelmRelease(context.TODO(), cfg, "", "", hrpSpec)
	if err != nil {
		log.Fatal(err)
	}
	println(rel.Name)
}
