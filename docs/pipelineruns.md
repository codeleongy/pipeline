<!--
---
linkTitle: "PipelineRuns"
weight: 500
---
-->
# PipelineRuns

<!-- toc -->
- [PipelineRuns](#pipelineruns)
  - [Overview](#overview)
  - [Configuring a <code>PipelineRun</code>](#configuring-a-pipelinerun)
    - [Specifying the target <code>Pipeline</code>](#specifying-the-target-pipeline)
      - [Tekton Bundles](#tekton-bundles)
      - [Remote Pipelines](#remote-pipelines)
    - [Specifying <code>Resources</code>](#specifying-resources)
    - [Specifying Task-level `ComputeResources`](#specifying-task-level-computeresources)
    - [Specifying <code>Parameters</code>](#specifying-parameters)
      - [Propagated Parameters](#propagated-parameters)
        - [Scope and Precedence](#scope-and-precedence)
        - [Default Values](#default-values)
        - [Referenced Resources](#referenced-resources)
        - [Object Parameters](#object-parameters) 
    - [Specifying custom <code>ServiceAccount</code> credentials](#specifying-custom-serviceaccount-credentials)
    - [Mapping <code>ServiceAccount</code> credentials to <code>Tasks</code>](#mapping-serviceaccount-credentials-to-tasks)
    - [Specifying a <code>Pod</code> template](#specifying-a-pod-template)
    - [Specifying taskRunSpecs](#specifying-taskrunspecs)
    - [Specifying <code>Workspaces</code>](#specifying-workspaces)
      - [Propagated Workspaces](#propagated-workspaces)
        - [Referenced Resources](#workspace-referenced-resources)
        - [Referenced TaskRuns within Embedded PipelineRuns](#referenced-taskruns-within-embedded-pipelineruns)
    - [Specifying <code>LimitRange</code> values](#specifying-limitrange-values)
    - [Configuring a failure timeout](#configuring-a-failure-timeout)
  - [<code>PipelineRun</code> status](#pipelinerun-status)
    - [The <code>status</code> field](#the-status-field) 
    - [Configuring usage of <code>TaskRun</code> and <code>Run</code> embedded statuses](#configuring-usage-of-taskrun-and-run-embedded-statuses)
    - [Monitoring execution status](#monitoring-execution-status)
  - [Cancelling a <code>PipelineRun</code>](#cancelling-a-pipelinerun)
  - [Gracefully cancelling a <code>PipelineRun</code>](#gracefully-cancelling-a-pipelinerun)
  - [Gracefully stopping a <code>PipelineRun</code>](#gracefully-stopping-a-pipelinerun)
  - [Pending <code>PipelineRuns</code>](#pending-pipelineruns)
<!-- /toc -->


## Overview

A `PipelineRun` allows you to instantiate and execute a [`Pipeline`](pipelines.md) on-cluster.
A `Pipeline` specifies one or more `Tasks` in the desired order of execution. A `PipelineRun`
executes the `Tasks` in the `Pipeline` in the order they are specified until all `Tasks` have
executed successfully or a failure occurs.

**Note:** A `PipelineRun` automatically creates corresponding `TaskRuns` for every
`Task` in your `Pipeline`.

The `Status` field tracks the current state of a `PipelineRun`, and can be used to monitor
progress.
This field contains the status of every `TaskRun`, as well as the full `PipelineSpec` used
to instantiate this `PipelineRun`, for full auditability.

## Configuring a `PipelineRun`

A `PipelineRun` definition supports the following fields:

- Required:
  - [`apiVersion`][kubernetes-overview] - Specifies the API version. For example
    `tekton.dev/v1beta1`.
  - [`kind`][kubernetes-overview] - Indicates that this resource object is a `PipelineRun` object.
  - [`metadata`][kubernetes-overview] - Specifies the metadata that uniquely identifies the
    `PipelineRun` object. For example, a `name`.
  - [`spec`][kubernetes-overview] - Specifies the configuration information for
    this `PipelineRun` object.
    - [`pipelineRef` or `pipelineSpec`](#specifying-the-target-pipeline) - Specifies the target [`Pipeline`](pipelines.md).
- Optional:
  - [`resources`](#specifying-resources) - Specifies the [`PipelineResources`](resources.md) to provision
    for executing the target `Pipeline`.
  - [`params`](#specifying-parameters) - Specifies the desired execution parameters for the `Pipeline`.
  - [`serviceAccountName`](#specifying-custom-serviceaccount-credentials) - Specifies a `ServiceAccount`
    object that supplies specific execution credentials for the `Pipeline`.
  - [`status`](#cancelling-a-pipelinerun) - Specifies options for cancelling a `PipelineRun`. 
  - [`taskRunSpecs`](#specifying-taskrunspecs) - Specifies a list of `PipelineRunTaskSpec` which allows for setting `ServiceAccountName`, [`Pod` template](./podtemplates.md), and `Metadata` for each task. This overrides the `Pod` template set for the entire `Pipeline`.
  - [`timeout`](#configuring-a-failure-timeout) - Specifies the timeout before the `PipelineRun` fails. `timeout` is deprecated and will eventually be removed, so consider using `timeouts` instead.
  - [`timeouts`](#configuring-a-failure-timeout) - Specifies the timeout before the `PipelineRun` fails. `timeouts` allows more granular timeout configuration, at the pipeline, tasks, and finally levels
  - [`podTemplate`](#specifying-a-pod-template) - Specifies a [`Pod` template](./podtemplates.md) to use as the basis for the configuration of the `Pod` that executes each `Task`.
  - [`workspaces`](#specifying-workspaces) - Specifies a set of workspace bindings which must match the names of workspaces declared in the pipeline being used. 

[kubernetes-overview]:
  https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

### Specifying the target `Pipeline`

You must specify the target `Pipeline` that you want the `PipelineRun` to execute, either by referencing
an existing `Pipeline` definition, or embedding a `Pipeline` definition directly in the `PipelineRun`.

To specify the target `Pipeline` by reference, use the `pipelineRef` field:

```yaml
spec:
  pipelineRef:
    name: mypipeline
```
To embed a `Pipeline` definition in the `PipelineRun`, use the `pipelineSpec` field:

```yaml
spec:
  pipelineSpec:
    tasks:
      - name: task1
        taskRef:
          name: mytask
```

The `Pipeline` in the [`pipelineSpec` example](../examples/v1beta1/pipelineruns/pipelinerun-with-pipelinespec.yaml)
example displays morning and evening greetings. Once you create and execute it, you can check the logs for its `Pods`:

```bash
kubectl logs $(kubectl get pods -o name | grep pipelinerun-echo-greetings-echo-good-morning)
Good Morning, Bob!

kubectl logs $(kubectl get pods -o name | grep pipelinerun-echo-greetings-echo-good-night)
Good Night, Bob!
```

You can also embed a `Task` definition the embedded `Pipeline` definition:

```yaml
spec:
  pipelineSpec:
    tasks:
      - name: task1
        taskSpec:
          steps: ...
```

In the [`taskSpec` in `pipelineSpec` example](../examples/v1beta1/pipelineruns/pipelinerun-with-pipelinespec-and-taskspec.yaml)
it's `Tasks` all the way down!

You can also specify labels and annotations with `taskSpec` which are propagated to each `taskRun` and then to the
respective pods. These labels can be used to identify and filter pods for further actions (such as collecting pod metrics,
and cleaning up completed pod with certain labels, etc) even being part of one single Pipeline.

```yaml
spec:
  pipelineSpec:
    tasks:
      - name: task1
        taskSpec:
          metadata:
            labels:
              pipeline-sdk-type: kfp
        # ...
      - name: task2
        taskSpec:
          metadata:
            labels:
              pipeline-sdk-type: tfx
        # ...
```

#### Tekton Bundles

**Note: This is only allowed if `enable-tekton-oci-bundles` is set to
`"true"` in the `feature-flags` configmap, see [`install.md`](./install.md#customizing-the-pipelines-controller-behavior)**

You may also use a `Tekton Bundle` to reference a pipeline defined remotely.

 ```yaml
 spec:
   pipelineRef:
     name: mypipeline
     bundle: docker.io/myrepo/mycatalog:v1.0
 ```

The syntax and caveats are similar to using `Tekton Bundles` for  `Task` references
in [Pipelines](pipelines.md#tekton-bundles) or [TaskRuns](taskruns.md#tekton-bundles).

`Tekton Bundles` may be constructed with any toolsets that produce valid OCI image artifacts
so long as the artifact adheres to the [contract](tekton-bundle-contracts.md).

#### Remote Pipelines

**([beta feature](https://github.com/tektoncd/pipeline/blob/main/docs/install.md#beta-features))**

A `pipelineRef` field may specify a Pipeline in a remote location such as git.
Support for specific types of remote will depend on the Resolvers your
cluster's operator has installed. For more information including a tutorial, please check [resolution docs](resolution.md). The below example demonstrates
referencing a Pipeline in git:

```yaml
spec:
  pipelineRef:
    resolver: git
    params:
    - name: url
      value: https://github.com/tektoncd/catalog.git
    - name: revision
      value: abc123
    - name: pathInRepo
      value: /pipeline/buildpacks/0.1/buildpacks.yaml
```

### Specifying `Resources`

> :warning: **`PipelineResources` are [deprecated](deprecations.md#deprecation-table).**
>
> Consider using replacement features instead. Read more in [documentation](migrating-v1alpha1-to-v1beta1.md#replacing-pipelineresources-with-tasks)
> and [TEP-0074](https://github.com/tektoncd/community/blob/main/teps/0074-deprecate-pipelineresources.md).

A `Pipeline` requires [`PipelineResources`](resources.md) to provide inputs and store outputs
for the `Tasks` that comprise it. You must provision those resources in the `resources` field
in the `spec` section of the `PipelineRun` definition.

A `Pipeline` may require you to provision a number of different resources. For example:

- When executing a `Pipeline` against a pull request, the triggering
  system must specify the commit-ish of a `git` resource.
- When executing a `Pipeline` manually against your own environment, you
  must provision your GitHub fork using the `git` resource; your image
  registry using the `image` resource; and your Kubernetes cluster using the
  `cluster` resource.

You can reference a `PipelineResources` using the `resourceRef` field:

```yaml
spec:
  resources:
    - name: source-repo
      resourceRef:
        name: skaffold-git
    - name: web-image
      resourceRef:
        name: skaffold-image-leeroy-web
    - name: app-image
      resourceRef:
        name: skaffold-image-leeroy-app
```

You can also embed a `PipelineResource` definition in the `PipelineRun` using the `resourceSpec` field:

```yaml
spec:
  resources:
    - name: source-repo
      resourceSpec:
        type: git
        params:
          - name: revision
            value: v0.32.0
          - name: url
            value: https://github.com/GoogleContainerTools/skaffold
    - name: web-image
      resourceSpec:
        type: image
        params:
          - name: url
            value: gcr.io/christiewilson-catfactory/leeroy-web
    - name: app-image
      resourceSpec:
        type: image
        params:
          - name: url
            value: gcr.io/christiewilson-catfactory/leeroy-app
```

**Note:** All `persistentVolumeClaims` specified within a `PipelineRun` are bound
until their respective `Pods` or the entire `PipelineRun` are deleted. This also applies
to all `persistentVolumeClaims` generated internally.

### Specifying Task-level `ComputeResources`

**([alpha only](https://github.com/tektoncd/pipeline/blob/main/docs/install.md#alpha-features))**

Task-level compute resources can be configured in `PipelineRun.TaskRunSpecs.ComputeResources` or `TaskRun.ComputeResources`.

e.g.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: pipeline
spec:
  tasks:
    - name: task
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pipelinerun 
spec:
  pipelineRef:
    name: pipeline
  taskRunSpecs:
    - pipelineTaskName: task
      computeResources:
        requests:
          cpu: 2
```

Further details and examples could be found in [Compute Resources in Tekton](https://github.com/tektoncd/pipeline/blob/main/docs/compute-resources.md).

### Specifying `Parameters`

(See also [Specifying Parameters in Tasks](tasks.md#specifying-parameters))

You can specify `Parameters` that you want to pass to the `Pipeline` during execution,
including different values of the same parameter for different `Tasks` in the `Pipeline`.

**Note:** You must specify all the `Parameters` that the `Pipeline` expects. Parameters
that have default values specified in Pipeline are not required to be provided by PipelineRun.

For example:

```yaml
spec:
  params:
    - name: pl-param-x
      value: "100"
    - name: pl-param-y
      value: "500"
```
You can pass in extra `Parameters` if needed depending on your use cases. An example use
case is when your CI system autogenerates `PipelineRuns` and it has `Parameters` it wants to
provide to all `PipelineRuns`. Because you can pass in extra `Parameters`, you don't have to
go through the complexity of checking each `Pipeline` and providing only the required params.

#### Propagated Parameters

When using an inlined spec, parameters from the parent `PipelineRun` will be
propagated to any inlined specs without needing to be explicitly defined. This
allows authors to simplify specs by automatically propagating top-level
parameters down to other inlined resources.

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: pr-echo-
spec:
  params:
    - name: HELLO
      value: "Hello World!"
    - name: BYE
      value: "Bye World!"
  pipelineSpec:
    tasks:
      - name: echo-hello
        taskSpec:
          steps:
            - name: echo
              image: ubuntu
              script: |
                #!/usr/bin/env bash
                echo "$(params.HELLO)"
      - name: echo-bye
        taskSpec:
          steps:
            - name: echo
              image: ubuntu
              script: |
                #!/usr/bin/env bash
                echo "$(params.BYE)"
```

On executing the pipeline run, the parameters will be interpolated during resolution.
The specifications are not mutated before storage and so it remains the same.
The status is updated.

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pr-echo-szzs9
  ...
spec:
  params:
  - name: HELLO
    value: Hello World!
  - name: BYE
    value: Bye World!
  pipelineSpec:
    tasks:
    - name: echo-hello
      taskSpec:
        steps:
        - image: ubuntu
          name: echo
          script: |
            #!/usr/bin/env bash
            echo "$(params.HELLO)"
    - name: echo-bye
      taskSpec:
        steps:
        - image: ubuntu
          name: echo
          script: |
            #!/usr/bin/env bash
            echo "$(params.BYE)"
status:
  conditions:
  - lastTransitionTime: "2022-04-07T12:34:58Z"
    message: 'Tasks Completed: 2 (Failed: 0, Canceled 0), Skipped: 0'
    reason: Succeeded
    status: "True"
    type: Succeeded
  pipelineSpec:
    ...
  taskRuns:
    pr-echo-szzs9-echo-hello:
      pipelineTaskName: echo-hello
      status:
        ...
        taskSpec:
          steps:
          - image: ubuntu
            name: echo
            resources: {}
            script: |
              #!/usr/bin/env bash
              echo "Hello World!"
    pr-echo-szzs9-echo-bye:
      pipelineTaskName: echo-bye
      status:
        ...
        taskSpec:
          steps:
          - image: ubuntu
            name: echo
            resources: {}
            script: |
              #!/usr/bin/env bash
              echo "Bye World!"
```

##### Scope and Precedence

When Parameters names conflict, the inner scope would take precedence as shown in this example:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: pr-echo-
spec:
  params:
  - name: HELLO
    value: "Hello World!"
  - name: BYE
    value: "Bye World!"
  pipelineSpec:
    tasks:
      - name: echo-hello
        params:
        - name: HELLO
          value: "Sasa World!"
        taskSpec:
          params:
            - name: HELLO
              type: string
          steps:
            - name: echo
              image: ubuntu
              script: |
                #!/usr/bin/env bash
                echo "$(params.HELLO)"
    ...
```

resolves to

```yaml
# Successful execution of the above PipelineRun
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pr-echo-szzs9
  ...
spec:
  ...
status:
  conditions:
    - lastTransitionTime: "2022-04-07T12:34:58Z"
      message: 'Tasks Completed: 2 (Failed: 0, Canceled 0), Skipped: 0'
      reason: Succeeded
      status: "True"
      type: Succeeded
  ...
  taskRuns:
    pr-echo-szzs9-echo-hello:
      pipelineTaskName: echo-hello
      status:
        conditions:
          - lastTransitionTime: "2022-04-07T12:34:57Z"
            message: All Steps have completed executing
            reason: Succeeded
            status: "True"
            type: Succeeded
        taskSpec:
          steps:
            - image: ubuntu
              name: echo
              resources: {}
              script: |
                #!/usr/bin/env bash
                echo "Sasa World!"
          ...
```

##### Default Values

When `Parameter` specifications have default values, the `Parameter` value provided at runtime would take precedence to give users control, as shown in this example:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: pr-echo-
spec:
  params:
  - name: HELLO
    value: "Hello World!"
  - name: BYE
    value: "Bye World!"
  pipelineSpec:
    tasks:
      - name: echo-hello
        taskSpec:
          params:
          - name: HELLO
            type: string
            default: "Sasa World!"
          steps:
            - name: echo
              image: ubuntu
              script: |
                #!/usr/bin/env bash
                echo "$(params.HELLO)"
    ...
```

resolves to

```yaml
# Successful execution of the above PipelineRun
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pr-echo-szzs9
  ...
spec:
  ...
status:
  conditions:
    - lastTransitionTime: "2022-04-07T12:34:58Z"
      message: 'Tasks Completed: 2 (Failed: 0, Canceled 0), Skipped: 0'
      reason: Succeeded
      status: "True"
      type: Succeeded
  ...
  taskRuns:
    pr-echo-szzs9-echo-hello:
      pipelineTaskName: echo-hello
      status:
        conditions:
          - lastTransitionTime: "2022-04-07T12:34:57Z"
            message: All Steps have completed executing
            reason: Succeeded
            status: "True"
            type: Succeeded
        taskSpec:
          steps:
            - image: ubuntu
              name: echo
              resources: {}
              script: |
                #!/usr/bin/env bash
                echo "Hello World!"
          ...
```

##### Referenced Resources

When a PipelineRun definition has referenced specifications but does not explicitly pass Parameters, the PipelineRun will be created but the execution will fail because of missing Parameters.

```yaml
# Invalid PipelineRun attempting to propagate Parameters to referenced Tasks
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: pr-echo-
spec:
  params:
  - name: HELLO
    value: "Hello World!"
  - name: BYE
    value: "Bye World!"
  pipelineSpec:
    tasks:
      - name: echo-hello
        taskRef:
          name: echo-hello
      - name: echo-bye
        taskRef:
          name: echo-bye
---
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: echo-hello
spec:
  steps:
    - name: echo
      image: ubuntu
      script: |
        #!/usr/bin/env bash
        echo "$(params.HELLO)"
---
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: echo-bye
spec:
  steps:
    - name: echo
      image: ubuntu
      script: |
        #!/usr/bin/env bash
        echo "$(params.BYE)"
```

Fails as follows:

```yaml
# Failed execution of the above PipelineRun
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pr-echo-24lmf
  ...
spec:
  params:
  - name: HELLO
    value: Hello World!
  - name: BYE
    value: Bye World!
  pipelineSpec:
    tasks:
    - name: echo-hello
      taskRef:
        kind: Task
        name: echo-hello
    - name: echo-bye
      taskRef:
        kind: Task
        name: echo-bye
status:
  conditions:
  - lastTransitionTime: "2022-04-07T20:24:51Z"
    message: 'invalid input params for task echo-hello: missing values for
              these params which have no default values: [HELLO]'
    reason: PipelineValidationFailed
    status: "False"
    type: Succeeded
  ...
```

##### Object Parameters

**([alpha only](https://github.com/tektoncd/pipeline/blob/main/docs/install.md#alpha-features))**

When using an inlined spec, object parameters from the parent `PipelineRun` will also be
propagated to any inlined specs without needing to be explicitly defined. This
allows authors to simplify specs by automatically propagating top-level
parameters down to other inlined resources.
When propagating object parameters, scope and precedence also holds as shown below.
 
```yaml
apiVersion: tekton.dev/v1beta1 
kind: PipelineRun              
metadata:
  generateName: pipelinerun-object-param-result 
spec:
  params:
    - name: gitrepo            
      value:                   
        url: abc.com           
        commit: sha123         
  pipelineSpec:                
    tasks:                     
      - name: task1            
        params:                
          - name: gitrepo      
            value:
              branch: main     
              url: xyz.com     
        taskSpec:
          steps:
            - name: write-result            
              image: bash      
              args: [          
                "echo",        
                "--url=$(params.gitrepo.url)",  
                "--commit=$(params.gitrepo.commit)",
                "--branch=$(params.gitrepo.branch)",
              ]      
```

resolves to

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pipelinerun-object-param-resultpxp59
  ...
spec:
  params:
  - name: gitrepo
    value:
      commit: sha123
      url: abc.com
  pipelineSpec:
    tasks:
    - name: task1
      params:
      - name: gitrepo
        value:
          branch: main
          url: xyz.com
      taskSpec:
        metadata: {}
        spec: null
        steps:
        - args:
          - echo
          - --url=$(params.gitrepo.url)
          - --commit=$(params.gitrepo.commit)
          - --branch=$(params.gitrepo.branch)
          image: bash
          name: write-result
          resources: {}
status:
  completionTime: "2022-09-08T17:22:01Z"
  conditions:
  - lastTransitionTime: "2022-09-08T17:22:01Z"
    message: 'Tasks Completed: 1 (Failed: 0, Cancelled 0), Skipped: 0'
    reason: Succeeded
    status: "True"
    type: Succeeded
  pipelineSpec:
    tasks:
    - name: task1
      params:
      - name: gitrepo
        value:
          branch: main
          url: xyz.com
      taskSpec:
        metadata: {}
        spec: null
        steps:
        - args:
          - echo
          - --url=xyz.com
          - --commit=sha123
          - --branch=main
          image: bash
          name: write-result
          resources: {}
  startTime: "2022-09-08T17:21:57Z"
  taskRuns:
    pipelinerun-object-param-resultpxp59-task1:
      pipelineTaskName: task1
      status:
        completionTime: "2022-09-08T17:22:01Z"
        conditions:
        - lastTransitionTime: "2022-09-08T17:22:01Z"
          message: All Steps have completed executing
          reason: Succeeded
          status: "True"
          type: Succeeded
        podName: pipelinerun-object-param-resultpxp59-task1-pod
        startTime: "2022-09-08T17:21:57Z"
        steps:
        - container: step-write-result
          ...
	taskSpec:
          steps:
          - args:
            - echo
            - --url=xyz.com
            - --commit=sha123
            - --branch=main
            image: bash
            name: write-result
            resources: {}
```

### Specifying custom `ServiceAccount` credentials

You can execute the `Pipeline` in your `PipelineRun` with a specific set of credentials by
specifying a `ServiceAccount` object name in the `serviceAccountName` field in your `PipelineRun`
definition. If you do not explicitly specify this, the `TaskRuns` created by your `PipelineRun`
will execute with the credentials specified in the `configmap-defaults` `ConfigMap`. If this
default is not specified, the `TaskRuns` will execute with the [`default` service account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#use-the-default-service-account-to-access-the-api-server)
set for the target [`namespace`](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/).

For more information, see [`ServiceAccount`](auth.md).

[`Custom tasks`](pipelines.md#using-custom-tasks) may or may not use a service account name.
Consult the documentation of the custom task that you are using to determine whether it supports a service account name.

### Mapping `ServiceAccount` credentials to `Tasks`

If you require more granularity in specifying execution credentials, use the `taskRunSpecs[].taskServiceAccountName` field to
map a specific `serviceAccountName` value to a specific `Task` in the `Pipeline`. This overrides the global
`serviceAccountName` you may have set for the `Pipeline` as described in the previous section.

For example, if you specify these mappings:

```yaml
spec:
  serviceAccountName: sa-1
  taskRunSpecs:
    - pipelineTaskName: build-task
      taskServiceAccountName: sa-for-build
```

for this `Pipeline`:

```yaml
kind: Pipeline
spec:
  tasks:
    - name: build-task
      taskRef:
        name: build-push
    - name: test-task
      taskRef:
        name: test
```

then `test-task` will execute using the `sa-1` account while `build-task` will execute with `sa-for-build`.

### Specifying a `Pod` template

You can specify a [`Pod` template](podtemplates.md) configuration that will serve as the configuration starting
point for the `Pod` in which the container images specified in your `Tasks` will execute. This allows you to
customize the `Pod` configuration specifically for each `TaskRun`.

In the following example, the `Task` defines a `volumeMount` object named `my-cache`. The `PipelineRun`
provisions this object for the `Task` using a `persistentVolumeClaim` and executes it as user 1001.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: mytask
spec:
  steps:
    - name: writesomething
      image: ubuntu
      command: ["bash", "-c"]
      args: ["echo 'foo' > /my-cache/bar"]
      volumeMounts:
        - name: my-cache
          mountPath: /my-cache
---
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: mypipeline
spec:
  tasks:
    - name: task1
      taskRef:
        name: mytask
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: mypipelinerun
spec:
  pipelineRef:
    name: mypipeline
  podTemplate:
    securityContext:
      runAsNonRoot: true
      runAsUser: 1001
    volumes:
      - name: my-cache
        persistentVolumeClaim:
          claimName: my-volume-claim
```

[`Custom tasks`](pipelines.md#using-custom-tasks) may or may not use a pod template.
Consult the documentation of the custom task that you are using to determine whether it supports a pod template.

### Specifying taskRunSpecs

Specifies a list of `PipelineTaskRunSpec` which contains `TaskServiceAccountName`, `TaskPodTemplate`
and `PipelineTaskName`. Mapping the specs to the corresponding `Task` based upon the `TaskName` a PipelineTask
will run with the configured  `TaskServiceAccountName` and `TaskPodTemplate` overwriting the pipeline
wide `ServiceAccountName`  and [`podTemplate`](./podtemplates.md) configuration,
for example:

```yaml
spec:
  podTemplate:
    securityContext:
      runAsUser: 1000
      runAsGroup: 2000
      fsGroup: 3000
  taskRunSpecs:
    - pipelineTaskName: build-task
      taskServiceAccountName: sa-for-build
      taskPodTemplate:
        nodeSelector:
          disktype: ssd
```

If used with this `Pipeline`,  `build-task` will use the task specific `PodTemplate` (where `nodeSelector` has `disktype` equal to `ssd`).
`PipelineTaskRunSpec` may also contain `StepOverrides` and `SidecarOverrides`; see
[Overriding `Task` `Steps` and `Sidecars`](./taskruns.md#overriding-task-steps-and-sidecars) for more information.

The optional annotations and labels can be added under a `Metadata` field as for a specific running context.

e.g.

Rendering needed secrets with Vault:

```yaml
spec:
  pipelineRef:
    name: pipeline-name
  taskRunSpecs:
    - pipelineTaskName: task-name
      metadata: 
        annotations:
          vault.hashicorp.com/agent-inject-secret-foo: "/path/to/foo"
          vault.hashicorp.com/role: role-name
```

Updating labels applied in a runtime context:

```yaml
spec:
  pipelineRef:
    name: pipeline-name
  taskRunSpecs:
    - pipelineTaskName: task-name
      metadata: 
        labels:
          app: cloudevent
```

If a metadata key is present in different levels, the value that will be used in the `PipelineRun` is determined using this precedence order: `PipelineRun.spec.taskRunSpec.metadata` > `PipelineRun.metadata` > `Pipeline.spec.tasks.taskSpec.metadata`.

### Specifying `Workspaces`

If your `Pipeline` specifies one or more `Workspaces`, you must map those `Workspaces` to
the corresponding physical volumes in your `PipelineRun` definition. For example, you
can map a `PersistentVolumeClaim` volume to a `Workspace` as follows:

```yaml
workspaces:
  - name: myworkspace # must match workspace name in Task
    persistentVolumeClaim:
      claimName: mypvc # this PVC must already exist
    subPath: my-subdir
```

For more information, see the following topics:
- For information on mapping `Workspaces` to `Volumes`, see [Specifying `Workspaces` in `PipelineRuns`](workspaces.md#specifying-workspaces-in-pipelineruns).
- For a list of supported `Volume` types, see [Specifying `VolumeSources` in `Workspaces`](workspaces.md#specifying-volumesources-in-workspaces).
- For an end-to-end example, see [`Workspaces` in a `PipelineRun`](../examples/v1beta1/pipelineruns/workspaces.yaml).

[`Custom tasks`](pipelines.md#using-custom-tasks) may or may not use workspaces.
Consult the documentation of the custom task that you are using to determine whether it supports workspaces.

#### Propagated Workspaces

**[beta](https://github.com/tektoncd/pipeline/blob/main/docs/install.md#beta-features))**

When using an embedded spec, workspaces from the parent `PipelineRun` will be
propagated to any inlined specs without needing to be explicitly defined. This
allows authors to simplify specs by automatically propagating top-level
workspaces down to other inlined resources.
**Workspace substutions will only be made for `commands`, `args` and `script` fields of `steps`, `stepTemplates`, and `sidecars`.**

```yaml
# Inline specifications of a PipelineRun
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: recipe-time-
spec:
  workspaces:
    - name: shared-data
      volumeClaimTemplate:
        spec:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 16Mi
          volumeMode: Filesystem
  pipelineSpec:
    #workspaces:
    #  - name: shared-data
    tasks:
    - name: fetch-secure-data
      # workspaces:
      #   - name: shared-data 
      taskSpec:
        # workspaces:
        #   - name: shared-data 
        steps:
        - name: fetch-and-write-secure
          image: ubuntu
          script: |
            echo hi >> $(workspaces.shared-data.path)/recipe.txt
    - name: print-the-recipe
      # workspaces:
      #   - name: shared-data 
      runAfter:
        - fetch-secure-data
      taskSpec:
        # workspaces:
        #   - name: shared-data 
        steps:
        - name: print-secrets
          image: ubuntu
          script: cat $(workspaces.shared-data.path)/recipe.txt
```

On executing the pipeline run, the workspaces will be interpolated during resolution.

```yaml
# Successful execution of the above PipelineRun
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: recipe-time-
  ...
spec:
  pipelineSpec:
  ...
status:
  completionTime: "2022-06-02T18:17:02Z"
  conditions:
  - lastTransitionTime: "2022-06-02T18:17:02Z"
    message: 'Tasks Completed: 2 (Failed: 0, Canceled 0), Skipped: 0'
    reason: Succeeded
    status: "True"
    type: Succeeded
  pipelineSpec:
    ...
  taskRuns:
    recipe-time-lslt9-fetch-secure-data:
      pipelineTaskName: fetch-secure-data
      status:
        ...
        taskSpec:
          steps:
          - image: ubuntu
            name: fetch-and-write-secure
            resources: {}
            script: |
              echo hi >> cat /workspace/shared-data/recipe.txt
          workspaces:
          - name: shared-data
    recipe-time-lslt9-print-the-recipe:
      pipelineTaskName: print-the-recipe
      status:
        ...
        taskSpec:
          steps:
          - image: ubuntu
            name: print-secrets
            resources: {}
            script: cat /workspace/shared-data/recipe.txt
          workspaces:
          - name: shared-data
```

##### Workspace Referenced Resources

`Workspaces` cannot be propagated to referenced specifications. For example, the following Pipeline will fail when executed because the workspaces defined in the PipelineRun cannot be propagated to the referenced Pipeline.

```yaml
# PipelineRun attempting to propagate Workspaces to referenced Tasks
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: shared-task-storage
spec:
  resources:
    requests:
      storage: 16Mi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
---
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: fetch-and-print-recipe
spec:
  tasks:
  - name: fetch-the-recipe
    taskRef:
      name: fetch-secure-data
  - name: print-the-recipe
    taskRef:
      name: print-data
    runAfter:
      - fetch-the-recipe
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: recipe-time-
spec:
  pipelineRef:
    name: fetch-and-print-recipe
  workspaces:
  - name: shared-data
    persistentVolumeClaim:
      claimName: shared-task-storage
```

Upon execution, this will cause failures:

```yaml
# Failed execution of the above PipelineRun

apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: recipe-time-
  ...
spec:
  pipelineRef:
    name: fetch-and-print-recipe
  workspaces:
  - name: shared-data
    persistentVolumeClaim:
      claimName: shared-task-storage
status:
  completionTime: "2022-06-02T19:02:58Z"
  conditions:
  - lastTransitionTime: "2022-06-02T19:02:58Z"
    message: 'Tasks Completed: 1 (Failed: 1, Canceled 0), Skipped: 1'
    reason: Failed
    status: "False"
    type: Succeeded
  pipelineSpec:
    ...
  taskRuns:
    recipe-time-v5scg-fetch-the-recipe:
      pipelineTaskName: fetch-the-recipe
      status:
        completionTime: "2022-06-02T19:02:58Z"
        conditions:
        - lastTransitionTime: "2022-06-02T19:02:58Z"
          message: |
            "step-fetch-and-write" exited with code 1 (image: "docker.io/library/ubuntu@sha256:26c68657ccce2cb0a31b330cb0be2b5e108d467f641c62e13ab40cbec258c68d"); for logs run: kubectl -n default logs recipe-time-v5scg-fetch-the-recipe-pod -c step-fetch-and-write
          reason: Failed
          status: "False"
          type: Succeeded
        ...
        taskSpec:
          steps:
          - image: ubuntu
            name: fetch-and-write
            resources: {}
            script: | # See below: Replacements do not happen.
      	      echo hi >> $(workspaces.shared-data.path)/recipe.txt
```

#### Referenced TaskRuns within Embedded PipelineRuns
As mentioned in the [Workspace Referenced Resources](#workspace-referenced-resources), workspaces can only be propagated from PipelineRuns to embedded Pipeline specs, not Pipeline references. Similarly, workspaces can only be propagated from a Pipeline to embedded Task specs, not referenced Tasks. For example:

```yaml
# PipelineRun attempting to propagate Workspaces to referenced Tasks
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: fetch-secure-data
spec:
  workspaces: # If Referenced, Workspaces need to be explicitly declared
  - name: shared-data
  steps:
  - name: fetch-and-write
    image: ubuntu
    script: |
      echo $(workspaces.shared-data.path)      
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: recipe-time-
spec:
  workspaces:
  - name: shared-data
    persistentVolumeClaim:
      claimName: shared-task-storage
  pipelineSpec:
    # workspaces: # Since this is embedded specs, Workspaces don’t need to be declared
    #    ...
    tasks:
    - name: fetch-the-recipe
      workspaces: # If referencing resources, Workspaces need to be explicitly declared
      - name: shared-data
      taskRef: # Referencing a resource
        name: fetch-secure-data
    - name: print-the-recipe
      # workspaces: # Since this is embedded specs, Workspaces don’t need to be declared
      #    ...
      taskSpec:
        # workspaces: # Since this is embedded specs, Workspaces don’t need to be declared
        #    ...
        steps:
        - name: print-secrets
          image: ubuntu
          script: cat $(workspaces.shared-data.path)/recipe.txt
      runAfter:
        - fetch-the-recipe
```

The above pipelinerun successfully resolves to:

```yaml
# Successful execution of the above PipelineRun
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: recipe-time-
  ...
spec:
  pipelineSpec:
    ...
  workspaces:
  - name: shared-data
    persistentVolumeClaim:
      claimName: shared-task-storage
status:
  completionTime: "2022-06-09T18:42:14Z"
  conditions:
  - lastTransitionTime: "2022-06-09T18:42:14Z"
    message: 'Tasks Completed: 2 (Failed: 0, Cancelled 0), Skipped: 0'
    reason: Succeeded
    status: "True"
    type: Succeeded
  pipelineSpec:
    ...
  taskRuns:
    recipe-time-pj6l7-fetch-the-recipe:
      pipelineTaskName: fetch-the-recipe
      status:
        ...
        taskSpec:
          steps:
          - image: ubuntu
            name: fetch-and-write
            resources: {}
            script: |
              echo /workspace/shared-data
          workspaces:
          - name: shared-data
    recipe-time-pj6l7-print-the-recipe:
      pipelineTaskName: print-the-recipe
      status:
       ...
        taskSpec:
          steps:
          - image: ubuntu
            name: print-secrets
            resources: {}
            script: cat /workspace/shared-data/recipe.txt
          workspaces:
          - name: shared-data
```

### Specifying `LimitRange` values

In order to only consume the bare minimum amount of resources needed to execute one `Step` at a
time from the invoked `Task`, Tekton will request the compute values for CPU, memory, and ephemeral
storage for each `Step` based on the [`LimitRange`](https://kubernetes.io/docs/concepts/policy/limit-range/)
object(s), if present. Any `Request` or `Limit` specified by the user (on `Task` for example) will be left unchanged.

For more information, see the [`LimitRange` support in Pipeline](./compute-resources.md#limitrange-support).

### Configuring a failure timeout

You can use the `timeouts` field to set the `PipelineRun's` desired timeout value in minutes.  There are three sub-fields than can be used to specify failures timeout for the entire pipeline, for tasks, and for `finally` tasks.

```yaml
timeouts:
  pipeline: "0h0m60s"
  tasks: "0h0m40s"
  finally: "0h0m20s"
```
All three sub-fields are optional, and will be automatically processed according to the following constraint:
* `timeouts.pipeline >= timeouts.tasks + timeouts.finally`

Example timeouts usages are as follows:

Combination 1: Set the timeout for the entire `pipeline` and reserve a portion of it for `tasks`.

```yaml
kind: PipelineRun
spec:
  timeouts:
    pipeline: "0h4m0s"
    tasks: "0h1m0s"
```

Combination 2: Set the timeout for the entire `pipeline` and reserve a portion of it for `finally`.

```yaml
kind: PipelineRun
spec:
  timeouts:
    pipeline: "0h4m0s"
    finally: "0h3m0s"
```

Combination 3: Set only a `tasks` timeout, with no timeout for the entire `pipeline`.

```yaml
kind: PipelineRun
spec:
  timeouts:
    pipeline: "0"  # No timeout
    tasks: "0h3m0s"
```

Combination : Set only a `finally` timeout, with no timeout for the entire `pipeline`.

```yaml
kind: PipelineRun
spec:
  timeouts:
    pipeline: "0"  # No timeout
    finally: "0h3m0s"
```

You can also use the *Deprecated* `timeout` field to set the `PipelineRun's` desired timeout value in minutes.
If you do not specify this value in the `PipelineRun`, the global default timeout value applies.
If you set the timeout to 0, the `PipelineRun` fails immediately upon encountering an error.

> :warning: ** `timeout` is deprecated and will be removed in future versions. Consider using `timeouts` instead.

If you do not specify the `timeout` value or `timeouts.pipeline` in the `PipelineRun`, the global default timeout value applies.
If you set the `timeout` value or `timeouts.pipeline` to 0, the `PipelineRun` fails immediately upon encountering an error.
If `timeouts.tasks` or `timeouts.finally` is set to 0, `timeouts.pipeline` must also be set to 0.

The global default timeout is set to 60 minutes when you first install Tekton. You can set
a different global default timeout value using the `default-timeout-minutes` field in
[`config/config-defaults.yaml`](./../config/config-defaults.yaml).

The `timeout` value is a `duration` conforming to Go's
[`ParseDuration`](https://golang.org/pkg/time/#ParseDuration) format. For example, valid
values are `1h30m`, `1h`, `1m`, and `60s`. If you set the global timeout to 0, all `PipelineRuns`
that do not have an individual timeout set will fail immediately upon encountering an error.

## `PipelineRun` status

### The `status` field

Your `PipelineRun`'s `status` field can contain the following fields:

- Required:
  - `status` - Most relevant, `status.conditions`, which contains the latest observations of the `PipelineRun`'s state. [See here](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties) for information on typical status properties. 
  - `startTime` - The time at which the `PipelineRun` began executing, in [RFC3339](https://tools.ietf.org/html/rfc3339) format.
  - `completionTime` - The time at which the `PipelineRun` finished executing, in [RFC3339](https://tools.ietf.org/html/rfc3339) format.
  - [`pipelineSpec`](pipelines.md#configuring-a-pipeline) - The exact `PipelineSpec` used when starting the `PipelineRun`.
- Optional:
  - `taskRuns` - A map of `TaskRun` names to detailed information about the status of that `TaskRun`. This is deprecated and will be removed in favor of using `childReferences`.
  - `runs` - A map of custom task `Run` names to detailed information about the status of that `Run`. This is deprecated and will be removed in favor of using `childReferences`.
  - [`pipelineResults`](pipelines.md#emitting-results-from-a-pipeline) - Results emitted by this `PipelineRun`.
  - `skippedTasks` - A list of `Task`s which were skipped when running this `PipelineRun` due to [when expressions](pipelines.md#guard-task-execution-using-when-expressions), including the when expressions applying to the skipped task.
  - `childReferences` - A list of references to each `TaskRun` or `Run` in this `PipelineRun`, which can be used to look up the status of the underlying `TaskRun` or `Run`. Each entry contains the following:
    - [`kind`][kubernetes-overview] - Generally either `TaskRun` or `Run`.
    - [`apiVersion`][kubernetes-overview] - The API version for the underlying `TaskRun` or `Run`.
    - [`whenExpressions`](pipelines.md#guard-task-execution-using-when-expressions) - The list of when expressions guarding the execution of this task.
  - `provenance` - Metadata about resources used in the PipelineRun such as the source from where a remote pipeline definition was fetched.
  - `finallyStartTime`- The time at which the PipelineRun's `finally` Tasks, if any, began
  executing, in [RFC3339](https://tools.ietf.org/html/rfc3339) format.

### Configuring usage of `TaskRun` and `Run` embedded statuses

Currently, the default behavior is to pupulate `status.childReferences` with references to the `TaskRun`s and
`Run`s, which can be used to look up their statuses.

This behavior can be controlled by changing the `embedded-status` feature flag in the `feature-flags`
config map. See [`install.md`](./install.md#customizing-the-pipelines-controller-behavior) for more
information on feature flags. The possible values for `embedded-status` are:
- `minimal` - The current default behavior, populate `status.childReferences`, not `status.taskRuns` or `status.runs`.
- `full` - Populating `status.taskRuns` and `status.runs`, without populating `status.childReferences`.
- `both` - Populate `status.childReferences` as well as `status.taskRuns` and `status.runs`.

*Note that after the `PipelineRunStatus` migration as planned in [TEP-100](https://github.com/tektoncd/community/blob/main/teps/0100-embedded-taskruns-and-runs-status-in-pipelineruns.md#2-deprecate-and-remove-full-embedded-status),
[the `full` and `both` `embedded-status` options will be removed](https://github.com/tektoncd/pipeline/blob/main/docs/deprecations.md).

### Monitoring execution status

As your `PipelineRun` executes, its `status` field accumulates information on the execution of each `TaskRun`
as well as the `PipelineRun` as a whole. This information includes the name of the pipeline `Task` associated
to a `TaskRun`, the complete [status of the `TaskRun`](taskruns.md#monitoring-execution-status) and details
about `whenExpressions` that may be associated to a `TaskRun`.

The following example shows an extract from the `status` field of a `PipelineRun` that has executed successfully:

```yaml
completionTime: "2020-05-04T02:19:14Z"
conditions:
  - lastTransitionTime: "2020-05-04T02:19:14Z"
    message: "Tasks Completed: 4, Skipped: 0"
    reason: Succeeded
    status: "True"
    type: Succeeded
startTime: "2020-05-04T02:00:11Z"
taskRuns:
  triggers-release-nightly-frwmw-build:
    pipelineTaskName: build
    status:
      completionTime: "2020-05-04T02:10:49Z"
      conditions:
        - lastTransitionTime: "2020-05-04T02:10:49Z"
          message: All Steps have completed executing
          reason: Succeeded
          status: "True"
          type: Succeeded
      podName: triggers-release-nightly-frwmw-build-pod
      resourcesResult:
        - key: commit
          resourceName: git-source-triggers-frwmw
          value: 9ab5a1234166a89db352afa28f499d596ebb48db
      startTime: "2020-05-04T02:05:07Z"
      steps:
        - container: step-build
          imageID: docker-pullable://golang@sha256:a90f2671330831830e229c3554ce118009681ef88af659cd98bfafd13d5594f9
          name: build
          terminated:
            containerID: docker://6b6471f501f59dbb7849f5cdde200f4eeb64302b862a27af68821a7fb2c25860
            exitCode: 0
            finishedAt: "2020-05-04T02:10:45Z"
            reason: Completed
            startedAt: "2020-05-04T02:06:24Z"
  ```

The following tables shows how to read the overall status of a `PipelineRun`.
Completion time is set once a `PipelineRun` reaches status `True` or `False`:

`status` | `reason`           | `completionTime` is set |                                                                           Description
:--------|:-------------------|:-----------------------:|-------------------------------------------------------------------------------------:
Unknown  | Started            |           No            |                          The `PipelineRun` has just been picked up by the controller.
Unknown  | Running            |           No            |                  The `PipelineRun` has been validate and started to perform its work.
Unknown  | Cancelled          |           No            | The user requested the PipelineRun to be cancelled. Cancellation has not be done yet.
True     | Succeeded          |           Yes           |                                             The `PipelineRun` completed successfully.
True     | Completed          |           Yes           |             The `PipelineRun` completed successfully, one or more Tasks were skipped.
False    | Failed             |           Yes           |                        The `PipelineRun` failed because one of the `TaskRuns` failed.
False    | \[Error message\]  |           Yes           |                 The `PipelineRun` failed with a permanent error (usually validation).
False    | Cancelled          |           Yes           |                                         The `PipelineRun` was cancelled successfully.
False    | PipelineRunTimeout |           Yes           |                                                          The `PipelineRun` timed out.

When a `PipelineRun` changes status, [events](events.md#pipelineruns) are triggered accordingly.

When a `PipelineRun` has `Tasks` that were `skipped`, the `reason` for skipping the task will be listed in the `Skipped Tasks` section of the `status` of the `PipelineRun`.

When a `PipelineRun` has `Tasks` with [`when` expressions](pipelines.md#guard-task-execution-using-when-expressions):
- If the `when` expressions evaluate to `true`, the `Task` is executed then the `TaskRun` and its resolved `when` expressions will be listed in the `Task Runs` section of the `status` of the `PipelineRun`.
- If the `when` expressions evaluate to `false`, the `Task` is skipped then its name and its resolved `when` expressions will be listed in the `Skipped Tasks` section of the `status` of the `PipelineRun`.

```yaml
Conditions:
  Last Transition Time:  2020-08-27T15:07:34Z
  Message:               Tasks Completed: 1 (Failed: 0, Cancelled 0), Skipped: 1
  Reason:                Completed
  Status:                True
  Type:                  Succeeded
Skipped Tasks:
  Name:       skip-this-task
  Reason:     When Expressions evaluated to false
  When Expressions:
    Input:     foo
    Operator:  in
    Values:
      bar
    Input:     foo
    Operator:  notin
    Values:
      foo
Task Runs:
  pipelinerun-to-skip-task-run-this-task:
    Pipeline Task Name:  run-this-task
    Status:
      ...
    When Expressions:
      Input:     foo
      Operator:  in
      Values:
        foo
```

The name of the `TaskRuns` and `Runs` owned by a `PipelineRun`  are univocally associated to the owning resource.
If a `PipelineRun` resource is deleted and created with the same name, the child `TaskRuns` will be created with the
same name as before. The base format of the name is `<pipelinerun-name>-<pipelinetask-name>`. If the `PipelineTask`
has a `Matrix`, the name will have an int suffix with format `<pipelinerun-name>-<pipelinetask-name>-<combination-id>`.
The name may vary according the logic of [`kmeta.ChildName`](https://pkg.go.dev/github.com/knative/pkg/kmeta#ChildName).

Some examples:

| `PipelineRun` Name                                       | `PipelineTask` Name                                          | `TaskRun` Names                                                                        |
|----------------------------------------------------------|--------------------------------------------------------------|----------------------------------------------------------------------------------------|
| pipeline-run                                             | task1                                                        | pipeline-run-task1                                                                     |
| pipeline-run                                             | task2-0123456789-0123456789-0123456789-0123456789-0123456789 | pipeline-runee4a397d6eab67777d4e6f9991cd19e6-task2-0123456789-0                        |
| pipeline-run-0123456789-0123456789-0123456789-0123456789 | task3                                                        | pipeline-run-0123456789-0123456789-0123456789-0123456789-task3                         |
| pipeline-run-0123456789-0123456789-0123456789-0123456789 | task2-0123456789-0123456789-0123456789-0123456789-0123456789 | pipeline-run-0123456789-012345607ad8c7aac5873cdfabe472a68996b5c                        |
| pipeline-run                                             | task4 (with 2x2 `Matrix`)                                    | pipeline-run-task1-0, pipeline-run-task1-2, pipeline-run-task1-3, pipeline-run-task1-4 |

## Cancelling a `PipelineRun`

To cancel a `PipelineRun` that's currently executing, update its definition
to mark it as "Cancelled". When you do so, the spawned `TaskRuns` are also marked
as cancelled, all associated `Pods` are deleted, and their `Retries` are not executed.
Pending `finally` tasks are not scheduled.

For example:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: go-example-git
spec:
  # […]
  status: "Cancelled"
```

## Gracefully cancelling a `PipelineRun`

To gracefully cancel a `PipelineRun` that's currently executing, update its definition
to mark it as "CancelledRunFinally". When you do so, the spawned `TaskRuns` are also marked
as cancelled, all associated `Pods` are deleted, and their `Retries` are not executed.
`finally` tasks are scheduled normally.

For example:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: go-example-git
spec:
  # […]
  status: "CancelledRunFinally"
```


## Gracefully stopping a `PipelineRun`

To gracefully stop a `PipelineRun` that's currently executing, update its definition
to mark it as "StoppedRunFinally". When you do so, the spawned `TaskRuns` are completed normally,
including executing their `retries`, but no new non-`finally` task is scheduled. `finally` tasks are executed afterwards.
For example:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: go-example-git
spec:
  # […]
  status: "StoppedRunFinally"
```

## Pending `PipelineRuns`

A `PipelineRun` can be created as a "pending" `PipelineRun` meaning that it will not actually be started until the pending status is cleared.

Note that a `PipelineRun` can only be marked "pending" before it has started, this setting is invalid after the `PipelineRun` has been started.

To mark a `PipelineRun` as pending, set `.spec.status` to `PipelineRunPending` when creating it:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: go-example-git
spec:
  # […]
  status: "PipelineRunPending"
```

To start the PipelineRun, clear the `.spec.status` field. Alternatively, update the value to `Cancelled` to cancel it.

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
