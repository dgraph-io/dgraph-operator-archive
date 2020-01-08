## dgraph-operator

Dgraph Operator creates/configures/manages Dgraph clusters atop Kubernetes.

### Synopsis

Dgraph Operator creates/configures/manages Dgraph clusters atop Kubernetes.

```
dgraph-operator [flags]
```

### Options

```
      --alsologtostderr                  log to standard error as well as files
      --config-file string               Configuration file. Takes precedence over default values, but is overridden to values set with environment variables and flags.
  -h, --help                             help for dgraph-operator
      --k8s-api-server-url string        URL of the kubernetes API server.
      --kubecfg-path string              Path of kubeconfig file.
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --server.host string               Host to listen on. (default "0.0.0.0")
      --server.port int                  Port to listen on. (default 7777)
      --skip-crd                         Skip kubernetes custom resource definition creation.
      --stderrthreshold severity         logs at or above this threshold go to stderr
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [dgraph-operator cmdref](dgraph-operator_cmdref.md)	 - Generate command line reference for dgraph operator command line interface.
* [dgraph-operator version](dgraph-operator_version.md)	 - Display the version of the current build of dgraph operator.

