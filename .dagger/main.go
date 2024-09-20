package main

import (
	"context"
	"dagger/demo/internal/dagger"
	"fmt"
	"strings"
	"time"
)

const (
	helmVersion     = "3.16.1"
	helmDocsVersion = "v1.14.2"
)

type Demo struct {
	// Project source directory
	//
	// +private
	Source *dagger.Directory
}

func New(
	// Project source directory.
	//
	// +defaultPath="/"
	// +ignore=[".devenv", ".direnv", ".github"]
	source *dagger.Directory,
) *Demo {
	return &Demo{
		Source: source,
	}
}

// Build the application container.
func (m *Demo) Build() *dagger.Container {
	return dag.Container().
		From("nginx:1.16.0").
		WithFile("/usr/share/nginx/html/index.html", m.Source.File("index.html"))
}

// Run the application (for demo purposes).
func (m *Demo) Serve() *dagger.Service {
	return m.Build().WithExposedPort(80).AsService()
}

// Test the Helm chart.
func (m *Demo) Test(ctx context.Context) error {
	app := m.Build()

	registry := dag.Registry().Service()

	// Push the container image to a local registry.
	_, err := dag.Container().From("quay.io/skopeo/stable").
		WithServiceBinding("registry", registry).
		WithMountedFile("/work/image.tar", app.AsTarball()).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"skopeo", "copy", "--all", "--dest-tls-verify=false", "docker-archive:/work/image.tar", "docker://registry:5000/demo-dagger-helm:latest"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	// Configure k3s to use the local registry.
	k8s := dag.K3S("test").With(func(k *dagger.K3S) *dagger.K3S {
		return k.WithContainer(
			k.Container().
				WithEnvVariable("BUST", time.Now().String()).
				WithExec([]string{"sh", "-c", `
cat <<EOF > /etc/rancher/k3s/registries.yaml
mirrors:
  "registry:5000":
    endpoint:
      - "http://registry:5000"
EOF`}).
				WithServiceBinding("registry", registry),
		)
	})

	// Start the Kubernetes cluster.
	_, err = k8s.Server().Start(ctx)
	if err != nil {
		return err
	}

	const values = `
image:
    repository: registry:5000/demo-dagger-helm
    tag: latest
`

	_, err = m.chart().Package().
		WithKubeconfigFile(k8s.Config()).
		Install("demo", dagger.HelmPackageInstallOpts{
			Wait: true,
			Values: []*dagger.File{
				dag.Directory().WithNewFile("values.yaml", values).File("values.yaml"),
			},
		}).
		Test(ctx, dagger.HelmReleaseTestOpts{
			Logs: true,
		})
	if err != nil {
		return err
	}

	return nil
}

// Lint the Helm chart.
func (m *Demo) Lint(ctx context.Context) (string, error) {
	chart := m.chart()

	return chart.Lint().Stdout(ctx)
}

// Package and release the Helm chart (and the application).
func (m *Demo) Release(ctx context.Context, version string, githubActor string, githubToken *dagger.Secret) error {
	_, err := m.Build().
		WithRegistryAuth("ghcr.io", githubActor, githubToken).
		Publish(ctx, fmt.Sprintf("ghcr.io/%s/demo-dagger-helm:%s", githubActor))
	if err != nil {
		return err
	}

	err = m.chart().
		Package(dagger.HelmChartPackageOpts{
			Version:    strings.TrimPrefix(version, "v"),
			AppVersion: version,
		}).
		WithRegistryAuth("ghcr.io", githubActor, githubToken).
		Publish(ctx, fmt.Sprintf("oci://ghcr.io/%s/helm-charts", githubActor))
	if err != nil {
		return err
	}

	return nil
}

func (m *Demo) chart() *dagger.HelmChart {
	chart := m.Source.Directory("deploy/charts/demo-dagger-helm")

	// Generate the README.md file using helm-docs.
	// See https://github.com/norwoodj/helm-docs
	readme := dag.HelmDocs(dagger.HelmDocsOpts{Version: helmDocsVersion}).Generate(chart)

	chart = chart.WithFile("README.md", readme)

	return dag.Helm(dagger.HelmOpts{Version: helmVersion}).Chart(chart)
}
