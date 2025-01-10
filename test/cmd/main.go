package main

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/databus23/helm-diff/v3/diff"
	"github.com/databus23/helm-diff/v3/manifest"

	addonsv1alpha1 "sigs.k8s.io/cluster-api-addon-provider-helm/api/v1alpha1"
	"sigs.k8s.io/cluster-api-addon-provider-helm/internal"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	cfg := config.GetConfigOrDie()
	println(cfg.Host)

	clnt := internal.HelmClient{}
	ctx := context.Background()
	hrpSpec := addonsv1alpha1.HelmReleaseProxySpec{
		ChartName:        "nginx-ingress",
		RepoURL:          "https://helm.nginx.com/stable",
		ReleaseName:      "my-release",
		ReleaseNamespace: "default",
		Values:           "",
	}
	release, err := clnt.GetHelmRelease(ctx, cfg, hrpSpec)
	if err != nil {
		log.Fatal(err)
	}
	println(release.Name)

	install, err := clnt.TemplateHelmRelease(ctx, cfg, "", "", hrpSpec)
	if err != nil {
		log.Fatal(err)
	}
	println(install.Name)
	_, actionConfig, err := internal.HelmInit(ctx, hrpSpec.ReleaseNamespace, cfg)
	if err != nil {
		log.Fatal(err)
	}

	original, err := actionConfig.KubeClient.Build(bytes.NewBuffer([]byte(release.Manifest)), false)
	if err != nil {
		log.Fatal(err)
	}
	target, err := actionConfig.KubeClient.Build(bytes.NewBuffer([]byte(install.Manifest)), false)
	if err != nil {
		log.Fatal(err)
	}

	releaseManifest, installManifest, err := manifest.Generate(original, target)
	if err != nil {
		log.Fatal(err)
	}
	currentSpecs := manifest.Parse(string(releaseManifest), hrpSpec.ReleaseNamespace, false, manifest.Helm3TestHook, manifest.Helm2TestSuccessHook)
	newSpecs := manifest.Parse(string(installManifest), hrpSpec.ReleaseNamespace, false, manifest.Helm3TestHook, manifest.Helm2TestSuccessHook)
	diffOptions := &diff.Options{}
	seenAnyChanges := diff.Manifests(currentSpecs, newSpecs, diffOptions, io.Discard)
	println(seenAnyChanges)
}
