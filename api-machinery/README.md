# API machinery

## Rest Mappping

Map a GroupVersionKind to HTTP REST endpoint in Kubernetes API Server and vice versa.

For example:

```bash
# /api/v1/namespaces/default/pods

# KindFor:
# -> GroupVersionKind: v1.Pod

# -> runtime.Object: &v1.Pod{}
```

## Scheme

Scheme is a registry maintaining a mapping of Kinds (strings) to Types (structs).

Schemes are dynamic - new types can be appended.

Let say we have an Go Struct for a CRD object:

```go
type Foo struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   FooSpec   `json:"spec"`
    Status FooStatus `json:"status"`
}
```

We can register the CRD object to the scheme by:

```go
scheme := runtime.NewScheme()
scheme.AddKnownTypes(schema.GroupVersion{
    Group:   "example.com",
    Version: "v1",
}, &Foo{})
```
