# estafette-ci-operator
The Estafette CI operator can perform upgrades to all components that together form the Estafette CI system and handle configuration through CRDs

## Installing the Estafette CI operator

To install use the following commands:

```bash
export NAMESPACE=estafette-ci
export VERSION=0.0.1
curl https://raw.githubusercontent.com/estafette/estafette-ci-operator/master/bundle.yaml | envsubst \$NAMESPACE,\$VERSION | kubectl apply -f -
```

## Installing Estafette CI via CRD

Once the operator is up and running you can install Estafette CI by creating the following CRD:

```yaml


```

## Configuring Estafette CI via CRD

