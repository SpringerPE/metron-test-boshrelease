Copied from https://github.com/cloudfoundry/loggregator/tree/master/router/internal
Because of the fact that these libs are in a `/internal` path, it cannot be reused
from other external projects (like this one)

```
An import of a path containing the element “internal” is disallowed if the
importing code is outside the tree rooted at the parent of the “internal” directory.
```

See https://golang.org/doc/go1.4#internalpackages
