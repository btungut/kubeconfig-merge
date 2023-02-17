# `kubeconfig-merge`: Painless and faultless way to merge kubeconfig files
---
![release](https://img.shields.io/github/v/release/btungut/kubeconfig-merge)
![Code Coverage](https://img.shields.io/badge/Code%20Coverage-51%25-yellow?style=flat)
![build](https://img.shields.io/github/actions/workflow/status/btungut/kubeconfig-merge/ci.yml?branch=master)
![go](https://img.shields.io/github/go-mod/go-version/btungut/kubeconfig-merge)

## Why is it developed?
If you are someone who works with a significant number of Kubernetes clusters, dealing with `kubecontext` in a manual way can be **boring** and also **result in problems**.
In addition to that, I am currently working with more than 20 customers, which results in an average of five clusters per customer.


## Arguments

| Argument     | Description                                                                | Default                                        |
|--------------|----------------------------------------------------------------------------|------------------------------------------------|
| file       | The additional kubeconfig file | *Required* |
| kubeconfig | The kubeconfig file which to be append into        | `KUBECONFIG` env variable, or `~/.kube/config` |
| name       | Context, cluster and user name of new entries                              | File name of `--file`|

## Examples

---
### `./kubeconfig-merge --file valid-default-cluster.yaml`

![kubeconfig-merge without name](.assets/kubeconfig-merge-01.png)

<br/>

---

### `./kubeconfig-merge --file valid-default-cluster.yaml --name foo`
![kubeconfig-merge with name](.assets/kubeconfig-merge-02.png)

## Contributors
[Ohki Nozomu](https://github.com/ohkinozomu)
 