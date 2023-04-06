# crossplane-crd-generator

this project permit to generate crossplane CRD from golang code.
To do that, it use the controller-gen (standard librairy used by operator-sdk) to generate kubernetes CRD from golang code. Then it patch the CRD generated to be used by crossplane.

## Installation

```
go install github.com/disaster37/crossplane-crd-generator@v1.0.1
```

## Usage

### Basic usage
```
crossplane-gen crd --source-path "./xrd/..." --target-path "crossplane"
```

### With advanced options
```bash
crossplane-gen crd --source-path "./xrd/..." --target-path "crossplane" --crd-options "generateEmbeddedObjectMeta=true" --crd-options "maxDescLen=0" --claim-name test --claim-plural-name tests
```

> You can also use options for `patchschema`. You can look available option on [controller-gen](https://book.kubebuilder.io/reference/controller-gen.html)