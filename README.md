# kubegodep2dep

The purpose of this tool is to generate (a part of) `Gopkg.toml` file that can be fed to the
[`dep`](https://golang.github.io/dep/) package manager in a project that wants to use
Kubernetes' libraries compatible with a particular Kubernetes version and among each other.

Managing constraints and/or overrides by hand is a huge PITA so why not generate them?

## Usage

1. Download `Godeps.json` of a particular Kubernetes version you want to use. Pick a tag or a branch.
For 1.12:
```console
curl -o Godeps.json https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.12/Godeps/Godeps.json
```
2. Install the binary from this repository:
```console
go get -u github.com/ash2k/kubegodep2dep
```
3. Run the tool:
```console
kubegodep2dep -godep ./Godeps.json > Gopkg-new.toml
```
4. Add other dependencies that your project needs and use `dep ensure` to pull them all down.
