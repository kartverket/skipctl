# ArgoKit v2

<p align="center">
<img src="logo.png" alt="ArgoKit Logo" width="300px" />
</p>

ArgoKit v2 is a set of reusable jsonnet templates, building on the old argoKit, that makes it easier to deploy
ArgoCD applications on SKIP. It is a work in progress and will be updated as we
learn more about how to best deploy with ArgoCD on SKIP. If you have any
questions, please reach out to the #gen-argocd or #spire-kartverket-devex-sparring channels in Slack.

## Installation

Assuming you have followed the [Getting Started](https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/554827836/Komme+i+gang+med+Argo+CD)
guide, you can use ArgoKit in your apps-repo.

First you need to install jsonnet bundler. This is done by running `brew install jsonnet-bundler` or download the
[binary release](https://github.com/jsonnet-bundler/jsonnet-bundler).
Next run `jb init`, and finally run `jb install github.com/kartverket/argokit@main`. If you want
to stay on a specific version use `jb install github.com/kartverket/argokit@2.0.0`

Remember to add `vendor/` to .gitignore

### Git submodules
Alternatively you can use submodules if you prefer it over jsonnet bundler (We recommend jsonnet bundler).
Include the ArgoKit library by running the following command:

```bash
git submodule add https://github.com/kartverket/argokit.git
```

### Automatic version updates

It is highly recommended to use dependabot to automatically update the ArgoKit
version when a new version is released. To do this, add the following to your
`.github/dependabot.yml` file:


```yaml
version: 2
updates:
  - package-ecosystem: gitsubmodule
    directory: /
    schedule:
      interval: daily
```

With this configuration, dependabot will automatically create a PR to update the
ArgoKit version once a day provided that a new version has been released.

## Usage

If you use jsonnet in your apps-repo, you can use the ArgoKit library to deploy
ArgoCD applications by including the `argokit.libsonnet` file in your jsonnet
file and calling the `argokit.Application` function. For example, to deploy an
application, you can use the following jsonnet file:

```jsonnet
local argokit = import 'github.com/kartverket/argokit/v2/jsonnet/argokit.libsonnet';
local application = argokit.appAndObjects.application;
application.new('foo-backend', "backend-image", 8080)
 + application.withReplicas(initial=2, max=5, targetCpuUtilization=80)
 + application.forHostnames('foo.bar.com')
 + application.withInboundSkipApp('foo-frontend')
```

## API Reference

[Read the full ArgoKit V2 API Reference here](./api-reference.md)

## Contributing

Contributions are welcome! Please open an issue or PR if you would like to
see something changed or added.

## License

ArgoKit is licensed under the MIT License. See [LICENSE](https://github.com/kartverket/argokit/blob/main/LICENSE) for the full
license text.
