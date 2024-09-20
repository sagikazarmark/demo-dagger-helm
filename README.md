# Demo: Testing and releasing Helm charts with [Dagger](https://dagger.io/)

This project demonstrates how to test and release Helm charts using [Dagger](https://dagger.io/).

The example chart for this demo is the default Helm chart created with `helm create`.

The actual command used to generate the chart is:

```sh
dagger -m github.com/sagikazarmark/daggerverse/helm@v0.13.0 call create --name demo-dagger-helm directory export --path deploy/charts/demo-dagger-helm --wipe
```

The Dagger module comes with the following commands:

- `build`: Build the application (nginx image with a custom `index.html`)
- `serve`: Serve the application (for demo purposes)
- `test`: Run Helm tests in a real Kubernetes cluster
- `lint`: Run `helm lint` on the chart
- `release`: Package and release the chart to a Helm repository

The repo also comes with a GitHub Actions workflow that runs the tests on every push and releases the chart on every tag.
