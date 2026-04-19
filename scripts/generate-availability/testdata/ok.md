---
title: Accessing contextual information about workflow runs
shortTitle: Contexts
intro: You can access context information in workflows and actions.
redirect_from:
  - /articles/contexts-and-expression-syntax-for-github-actions
  - /github/automating-your-workflow-with-github-actions/contexts-and-expression-syntax-for-github-actions
  - /actions/automating-your-workflow-with-github-actions/contexts-and-expression-syntax-for-github-actions
  - /actions/reference/contexts-and-expression-syntax-for-github-actions
  - /actions/reference/context-and-expression-syntax-for-github-actions
  - /actions/learn-github-actions/contexts
  - /actions/writing-workflows/choosing-what-your-workflow-does/contexts
versions:
  fpt: '*'
  ghes: '*'
  ghec: '*'
---

{% data reusables.actions.enterprise-github-hosted-runners %}

## About contexts

{% data reusables.actions.actions-contexts-about-description %} Each context is an object that contains properties, which can be strings or other objects.

{% data reusables.actions.context-contents %} For example, the `matrix` context is only populated for jobs in a [matrix](/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix).

You can access contexts using the expression syntax. For more information, see "[AUTOTITLE](/actions/learn-github-actions/expressions)."

{% raw %}
`${{ <context> }}`
{% endraw %}

{% data reusables.actions.context-injection-warning %}

| Context name | Type | Description |
|---------------|------|-------------|
| `github` | `object` | Information about the workflow run. For more information, see [`github` context](#github-context). |
| `env` | `object` | Contains variables set in a workflow, job, or step. For more information, see [`env` context](#env-context). |
| `vars` | `object` | Contains variables set at the repository, organization, or environment levels. For more information, see [`vars` context](#vars-context). |
| `job` | `object` | Information about the currently running job. For more information, see [`job` context](#job-context). |
| `jobs` | `object` | For reusable workflows only, contains outputs of jobs from the reusable workflow. For more information, see [`jobs` context](#jobs-context). |
| `steps` | `object` | Information about the steps that have been run in the current job. For more information, see [`steps` context](#steps-context). |
| `runner` | `object` | Information about the runner that is running the current job. For more information, see [`runner` context](#runner-context). |
| `secrets` | `object` | Contains the names and values of secrets that are available to a workflow run. For more information, see [`secrets` context](#secrets-context). |
| `strategy` | `object` | Information about the matrix execution strategy for the current job. For more information, see [`strategy` context](#strategy-context). |
| `matrix` | `object` | Contains the matrix properties defined in the workflow that apply to the current job. For more information, see [`matrix` context](#matrix-context). |
| `needs` | `object` | Contains the outputs of all jobs that are defined as a dependency of the current job. For more information, see [`needs` context](#needs-context). |
| `inputs` | `object` | Contains the inputs of a reusable or manually triggered workflow. For more information, see [`inputs` context](#inputs-context). |

As part of an expression, you can access context information using one of two syntaxes.

* Index syntax: `github['sha']`
* Property dereference syntax: `github.sha`

In order to use property dereference syntax, the property name must start with a letter or `_` and contain only alphanumeric characters, `-`, or `_`.

If you attempt to dereference a nonexistent property, it will evaluate to an empty string.

### Determining when to use contexts

{% data variables.product.prodname_actions %} includes a collection of variables called _contexts_ and a similar collection of variables called _default variables_. These variables are intended for use at different points in the workflow:

* **Default environment variables:** These environment variables exist only on the runner that is executing your job. For more information, see "[AUTOTITLE](/actions/learn-github-actions/variables#default-environment-variables)."
* **Contexts:** You can use most contexts at any point in your workflow, including when _default variables_ would be unavailable. For example, you can use contexts with expressions to perform initial processing before the job is routed to a runner for execution; this allows you to use a context with the conditional `if` keyword to determine whether a step should run. Once the job is running, you can also retrieve context variables from the runner that is executing the job, such as `runner.os`. For details of where you can use various contexts within a workflow, see "[Context availability](#context-availability)."

The following example demonstrates how these different types of variables can be used together in a job:

{% raw %}

```yaml copy
name: CI
on: push
jobs:
  prod-check:
    if: ${{ github.ref == 'refs/heads/main' }}
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploying to production server on branch $GITHUB_REF"
```

{% endraw %}

In this example, the `if` statement checks the [`github.ref`](/actions/learn-github-actions/contexts#github-context) context to determine the current branch name; if the name is `refs/heads/main`, then the subsequent steps are executed. The `if` check is processed by {% data variables.product.prodname_actions %}, and the job is only sent to the runner if the result is `true`. Once the job is sent to the runner, the step is executed and refers to the [`$GITHUB_REF`](/actions/learn-github-actions/variables#default-environment-variables) variable from the runner.

### Context availability

Different contexts are available throughout a workflow run. For example, the `secrets` context may only be used at certain places within a job.

In addition, some functions may only be used in certain places. For example, the `hashFiles` function is not available everywhere.

The following table indicates where each context and special function can be used within a workflow. Unless listed below, a function can be used anywhere.

| Workflow key | Context | Special functions |
| ---- | ------- | ----------------- |
| `run-name` | `github, inputs, vars` | None |
| `concurrency` | `github, inputs, vars` | None |
| `env` | `github, secrets, inputs, vars` | None |
| `jobs.<job_id>.concurrency` | `github, needs, strategy, matrix, inputs, vars` | None |
| `jobs.<job_id>.container` | `github, needs, strategy, matrix, vars, inputs` | None |
| `jobs.<job_id>.container.credentials` | `github, needs, strategy, matrix, env, vars, secrets, inputs` | None |
| `jobs.<job_id>.container.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, inputs` | None |
| `jobs.<job_id>.container.image` | `github, needs, strategy, matrix, vars, inputs` | None |
| `jobs.<job_id>.continue-on-error` | `github, needs, strategy, vars, matrix, inputs` | None |
| `jobs.<job_id>.defaults.run` | `github, needs, strategy, matrix, env, vars, inputs` | None |
| `jobs.<job_id>.env` | `github, needs, strategy, matrix, vars, secrets, inputs` | None |
| `jobs.<job_id>.environment` | `github, needs, strategy, matrix, vars, inputs` | None |
| `jobs.<job_id>.environment.url` | `github, needs, strategy, matrix, job, runner, env, vars, steps, inputs` | None |
| `jobs.<job_id>.if` | `github, needs, vars, inputs` | `always, cancelled, success, failure` |
| `jobs.<job_id>.name` | `github, needs, strategy, matrix, vars, inputs` | None |
| `jobs.<job_id>.outputs.<output_id>` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | None |
| `jobs.<job_id>.runs-on` | `github, needs, strategy, matrix, vars, inputs` | None |
| `jobs.<job_id>.secrets.<secrets_id>` | `github, needs, strategy, matrix, secrets, inputs, vars` | None |
| `jobs.<job_id>.services` | `github, needs, strategy, matrix, vars, inputs` | None |
| `jobs.<job_id>.services.<service_id>.credentials` | `github, needs, strategy, matrix, env, vars, secrets, inputs` | None |
| `jobs.<job_id>.services.<service_id>.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, inputs` | None |
| `jobs.<job_id>.steps.continue-on-error` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.env` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.if` | `github, needs, strategy, matrix, job, runner, env, vars, steps, inputs` | `always, cancelled, success, failure, hashFiles` |
| `jobs.<job_id>.steps.name` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.run` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.timeout-minutes` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.with` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.working-directory` | `github, needs, strategy, matrix, job, runner, env, vars, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.strategy` | `github, needs, vars, inputs` | None |
| `jobs.<job_id>.timeout-minutes` | `github, needs, strategy, matrix, vars, inputs` | None |
| `jobs.<job_id>.with.<with_id>` | `github, needs, strategy, matrix, inputs, vars` | None |
| `on.workflow_call.inputs.<inputs_id>.default` | `github, inputs, vars` | None |
| `on.workflow_call.outputs.<output_id>.value` | `github, jobs, vars, inputs` | None |

### Example: printing context information to the log

...
