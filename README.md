# Demo: Testing and releasing Helm charts with [Dagger](https://dagger.io/)

[![built with nix](https://builtwithnix.org/badge.svg)](https://builtwithnix.org)

This project demonstrates how to test and release Helm charts using [Dagger](https://dagger.io/).

The example chart for this demo is the default Helm chart created with `helm create`.

The actual command used to generate the chart is:

```sh
dagger -m github.com/sagikazarmark/daggerverse/helm@v0.13.0 call create --name demo directory export --path deploy/charts/demo --wipe
```

The Dagger module comes with the following commands:

- `build`: Build the application (nginx image with a custom `index.html`)
- `serve`: Serve the application (for demo purposes)
- `test`: Run Helm tests in a real Kubernetes cluster
- `lint`: Run `helm lint` on the chart
