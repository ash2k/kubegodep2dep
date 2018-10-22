# kubegodep2dep

The purpose of this tool is to generate (a part of) `Gopkg.toml` file that can be fed to the
[`dep`](https://golang.github.io/dep/) package manager in a project that wants to use
Kubernetes' libraries compatible with a particular Kubernetes version and among each other.

The huge benefit is that **exactly** the same dependency revisions are used. That means the set
of dependency versions that you get have been tested by Kubernetes unit, integration and
end-to-end tests plus real world usage so it is guaranteed to work.

Managing constraints and/or overrides by hand is a huge PITA so why not generate them?

## Usage
See the [compatibility-matrix](https://github.com/kubernetes/client-go#compatibility-matrix) to get an overview, which `client-go` version works with what Kubernetes release.

1. Install the binary from this repository:
    ```console
    go get -u github.com/ash2k/kubegodep2dep
    ```

1. Run the tool:
    ```console
    kubegodep2dep -kube-branch release-1.12 -client-go-branch release-9.0 > Gopkg-new.toml
    ```
    You can use `-godep` to pass another `Godep.json` other than the default file into the tool. URLs are supported as input, too.

1. Add other dependencies that your project needs and use `dep ensure` to pull them all down.
