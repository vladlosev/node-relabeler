# node-relabeler

[![vladlosev](https://circleci.com/gh/vladlosev/node-labeler.svg?style=svg)](https://circleci.com/gh/vladlosev/node-relabeler)

## About

This program watches for the changes in nodes of a Kubernetes cluster, and
sets labels on them according to specs provided to it.

The original impetus for this project came from desire to label Kubernetes
nodes with roles. If you label a node with the label
`node-role.kubernetes.io/<role>=`, running `kubectl get nodes` with display
the specified role in the `ROLE` column.

If you administer a Kubernetes cluster, it was possible to specify role
assignments in the kubelet startup options. However, starting with Kubernetes
1.16, specifying `node-role.kubernetes.io/<role>=` is no longer allowed. The
suggested workaround from the Kubernetes team is to label the nodes with some
other, allowed label and run a worker that would apply
`node-role.kubernetes.io/<role>=` to any nodes that are added or get updated
with that other label. This project is intended to provide such functionality.

## Running

When run, `node-relabeler` expects one or more re-labeling specifications
provided in the command line:
```
node-relabeler --relabel=old/label=old-value:new/label=new-value
```
(multiple `--relabel` options may be provided).  The program will watch for nodes
being added or updated with labels that match the `old-label=old-name` pattern
and if found, add the `new/label=new-value` label.  Either old label name or
value can contain a wildcard character `*` which will cause glob pattern matching.
If `*` also shows up in the new label name or value, it will be replaced with the
matched part of the old label.

For example,
- `--relabel=foo=*:bar=*` will add the label `bar` with the value of the
  existing label `foo`.
- `--relable=role=*:node-role.kubernetes.io/*=` will add the label
  `node-role.kubernetes.io/<role>=` with the value of the existing label
  `role`.

## Deploying

You can deploy `node-relabeler` into a Kubernetes cluster using a Helm chart
provided with the project in [charts/node-relabeler](charts/node-relabeler).

If you want to use Helm Operator, a sample Helm release object is available
in [deploy/helm-operator/release.yaml](deploy/helm-operator/release.yaml)
