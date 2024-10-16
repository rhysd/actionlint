---
title: Contexts
shortTitle: Contexts
intro: You can access context information in workflows and actions.
redirect_from:
  - /articles/contexts-and-expression-syntax-for-github-actions
  - /github/automating-your-workflow-with-github-actions/contexts-and-expression-syntax-for-github-actions
  - /actions/automating-your-workflow-with-github-actions/contexts-and-expression-syntax-for-github-actions
  - /actions/reference/contexts-and-expression-syntax-for-github-actions
  - /actions/reference/context-and-expression-syntax-for-github-actions
versions:
  fpt: '*'
  ghes: '*'
  ghae: '*'
  ghec: '*'
miniTocMaxHeadingLevel: 3
---

{% data reusables.actions.enterprise-beta %}
{% data reusables.actions.enterprise-github-hosted-runners %}

## About contexts

Contexts are a way to access information about workflow runs, runner environments, jobs, and steps. Each context is an object that contains properties, which can be strings or other objects.

{% data reusables.actions.context-contents %} For example, the `matrix` context is only populated for jobs in a [matrix](/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix).

You can access contexts using the expression syntax. For more information, see "[Expressions](/actions/learn-github-actions/expressions)."

{% raw %}
`${{ <context> }}`
{% endraw %}

{% data reusables.actions.context-injection-warning %}

| Context name | Type | Description |
|---------------|------|-------------|
| `github` | `object` | Information about the workflow run. For more information, see [`github` context](#github-context). |
| `env` | `object` | Contains environment variables set in a workflow, job, or step. For more information, see [`env` context](#env-context). |
| `job` | `object` | Information about the currently running job. For more information, see [`job` context](#job-context). |
{%- ifversion fpt or ghes > 3.3 or ghae-issue-4757 or ghec %}
| `jobs` | `object` | For reusable workflows only, contains outputs of jobs from the reusable workflow. For more information, see [`jobs` context](#jobs-context). |{% endif %}
| `steps` | `object` | Information about the steps that have been run in the current job. For more information, see [`steps` context](#steps-context). |
| `runner` | `object` | Information about the runner that is running the current job. For more information, see [`runner` context](#runner-context). |
| `secrets` | `object` | Contains the names and values of secrets that are available to a workflow run. For more information, see [`secrets` context](#secrets-context). |
| `strategy` | `object` | Information about the matrix execution strategy for the current job. For more information, see [`strategy` context](#strategy-context). |
| `matrix` | `object` | Contains the matrix properties defined in the workflow that apply to the current job. For more information, see [`matrix` context](#matrix-context). |
| `needs` | `object` | Contains the outputs of all jobs that are defined as a dependency of the current job. For more information, see [`needs` context](#needs-context). |
{%- ifversion fpt or ghec or ghes > 3.3 or ghae-issue-4757 %}
| `inputs` | `object` | Contains the inputs of a reusable {% ifversion actions-unified-inputs %}or manually triggered {% endif %}workflow. For more information, see [`inputs` context](#inputs-context). |{% endif %}

As part of an expression, you can access context information using one of two syntaxes.

- Index syntax: `github['sha']`
- Property dereference syntax: `github.sha`

In order to use property dereference syntax, the property name must start with a letter or `_` and contain only alphanumeric characters, `-`, or `_`.

If you attempt to dereference a non-existent property, it will evaluate to an empty string.

### Determining when to use contexts

{% data reusables.actions.using-context-or-environment-variables %}

### Context availability

Different contexts are available throughout a workflow run. For example, the `secrets` context may only be used at certain places within a job.

In addition, some functions may only be used in certain places. For example, the `hashFiles` function is not available everywhere.

The following table indicates where each context and special function can be used within a workflow. Unless listed below, a function can be used anywhere.

{% ifversion fpt or ghes > 3.3 or ghae-issue-4757 or ghec %}

| Workflow key | Context | Special functions |
| ---- | ------- | ----------------- |
| `concurrency` | `github, inputs` | None |
| `env` | `github, secrets, inputs` | None |
| `jobs.<job_id>.concurrency` | `github, needs, strategy, matrix, inputs` | None |
| `jobs.<job_id>.container` | `github, needs, strategy, matrix, env, secrets, inputs` | None |
| `jobs.<job_id>.container.credentials` | `github, needs, strategy, matrix, env, secrets, inputs` | None |
| `jobs.<job_id>.container.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, inputs` | None |
| `jobs.<job_id>.continue-on-error` | `github, needs, strategy, matrix, inputs` | None |
| `jobs.<job_id>.defaults.run` | `github, needs, strategy, matrix, env, inputs` | None |
| `jobs.<job_id>.env` | `github, needs, strategy, matrix, secrets, inputs` | None |
| `jobs.<job_id>.environment` | `github, needs, strategy, matrix, inputs` | None |
| `jobs.<job_id>.environment.url` | `github, needs, strategy, matrix, job, runner, env, steps, inputs` | None |
| `jobs.<job_id>.if` | `github, needs, inputs` | `always, cancelled, success, failure` |
| `jobs.<job_id>.name` | `github, needs, strategy, matrix, inputs` | None |
| `jobs.<job_id>.outputs.<output_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | None |
| `jobs.<job_id>.runs-on` | `github, needs, strategy, matrix, inputs` | None |
| `jobs.<job_id>.secrets.<secrets_id>` | `github, needs,{% ifversion actions-reusable-workflow-matrix %} strategy, matrix,{% endif %} secrets{% ifversion actions-unified-inputs %}, inputs{% endif %}` | None |
| `jobs.<job_id>.services` | `github, needs, strategy, matrix, inputs` | None |
| `jobs.<job_id>.services.<service_id>.credentials` | `github, needs, strategy, matrix, env, secrets, inputs` | None |
| `jobs.<job_id>.services.<service_id>.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, inputs` | None |
| `jobs.<job_id>.steps.continue-on-error` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.env` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.if` | `github, needs, strategy, matrix, job, runner, env, steps, inputs` | `always, cancelled, success, failure, hashFiles` |
| `jobs.<job_id>.steps.name` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.run` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.timeout-minutes` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.with` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.working-directory` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.strategy` | `github, needs, inputs` | None |
| `jobs.<job_id>.timeout-minutes` | `github, needs, strategy, matrix, inputs` | None |
| `jobs.<job_id>.with.<with_id>` | `github, needs{% ifversion actions-reusable-workflow-matrix %}, strategy, matrix{% endif %}{% ifversion actions-unified-inputs %}, inputs{% endif %}` | None |
| `on.workflow_call.inputs.<inputs_id>.default` | `github{% ifversion actions-unified-inputs %}, inputs{% endif %}` | None |
| `on.workflow_call.outputs.<output_id>.value` | `github, jobs, inputs` | None |
{% else %}
| Path | Context | Special functions |
| ---- | ------- | ----------------- |
| `concurrency` | `github` | None |
| `env` | `github, secrets` | None |
| `jobs.<job_id>.concurrency` | `github, needs, strategy, matrix` | None |
| `jobs.<job_id>.container` | `github, needs, strategy, matrix` | None |
| `jobs.<job_id>.container.credentials` | `github, needs, strategy, matrix, env, secrets` | None |
| `jobs.<job_id>.container.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets` | None |
| `jobs.<job_id>.continue-on-error` | `github, needs, strategy, matrix` | None |
| `jobs.<job_id>.defaults.run` | `github, needs, strategy, matrix, env` | None |
| `jobs.<job_id>.env` | `github, needs, strategy, matrix, secrets` | None |
| `jobs.<job_id>.environment` | `github, needs, strategy, matrix` | None |
| `jobs.<job_id>.environment.url` | `github, needs, strategy, matrix, job, runner, env, steps` | None |
| `jobs.<job_id>.if` | `github, needs` | `always, cancelled, success, failure` |
| `jobs.<job_id>.name` | `github, needs, strategy, matrix` | None |
| `jobs.<job_id>.outputs.<output_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | None |
| `jobs.<job_id>.runs-on` | `github, needs, strategy, matrix` | None |
| `jobs.<job_id>.services` | `github, needs, strategy, matrix` | None |
| `jobs.<job_id>.services.<service_id>.credentials` | `github, needs, strategy, matrix, env, secrets` | None |
| `jobs.<job_id>.services.<service_id>.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets` | None |
| `jobs.<job_id>.steps.continue-on-error` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.env` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.if` | `github, needs, strategy, matrix, job, runner, env, steps` | `always, cancelled, success, failure, hashFiles` |
| `jobs.<job_id>.steps.name` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.run` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.timeout-minutes` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.with` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.working-directory` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.strategy` | `github, needs` | None |
| `jobs.<job_id>.timeout-minutes` | `github, needs, strategy, matrix` | None |
{% endif %}

### Context availability 2

Different contexts are available throughout a workflow run. For example, the `secrets` context may only be used at certain places within a job.

In addition, some functions may only be used in certain places. For example, the `hashFiles` function is not available everywhere.

The following table indicates where each context and special function can be used within a workflow. Unless listed below, a function can be used anywhere.

{% ifversion fpt or ghes > 3.3 or ghae-issue-4757 or ghec %}

| Workflow key | Context | Special functions |
| ---- | ------- | ----------------- |
| `concurrency` | `github, inputs` | |
| `env` | `github, secrets, inputs` | |
| `jobs.<job_id>.concurrency` | `github, needs, strategy, matrix, inputs` | |
| `jobs.<job_id>.container` | `github, needs, strategy, matrix, env, secrets, inputs` | |
| `jobs.<job_id>.container.credentials` | `github, needs, strategy, matrix, env, secrets, inputs` | |
| `jobs.<job_id>.container.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, inputs` | |
| `jobs.<job_id>.continue-on-error` | `github, needs, strategy, matrix, inputs` | |
| `jobs.<job_id>.defaults.run` | `github, needs, strategy, matrix, env, inputs` | |
| `jobs.<job_id>.env` | `github, needs, strategy, matrix, secrets, inputs` | |
| `jobs.<job_id>.environment` | `github, needs, strategy, matrix, inputs` | |
| `jobs.<job_id>.environment.url` | `github, needs, strategy, matrix, job, runner, env, steps, inputs` | |
| `jobs.<job_id>.if` | `github, needs, inputs` | `always, cancelled, success, failure` |
| `jobs.<job_id>.name` | `github, needs, strategy, matrix, inputs` | |
| `jobs.<job_id>.outputs.<output_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | |
| `jobs.<job_id>.runs-on` | `github, needs, strategy, matrix, inputs` | |
| `jobs.<job_id>.secrets.<secrets_id>` | `github, needs,{% ifversion actions-reusable-workflow-matrix %} strategy, matrix,{% endif %} secrets{% ifversion actions-unified-inputs %}, inputs{% endif %}` | |
| `jobs.<job_id>.services` | `github, needs, strategy, matrix, inputs` | |
| `jobs.<job_id>.services.<service_id>.credentials` | `github, needs, strategy, matrix, env, secrets, inputs` | |
| `jobs.<job_id>.services.<service_id>.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, inputs` | |
| `jobs.<job_id>.steps.continue-on-error` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.env` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.if` | `github, needs, strategy, matrix, job, runner, env, steps, inputs` | `always, cancelled, success, failure, hashFiles` |
| `jobs.<job_id>.steps.name` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.run` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.timeout-minutes` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.with` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.steps.working-directory` | `github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs` | `hashFiles` |
| `jobs.<job_id>.strategy` | `github, needs, inputs` | |
| `jobs.<job_id>.timeout-minutes` | `github, needs, strategy, matrix, inputs` | |
| `jobs.<job_id>.with.<with_id>` | `github, needs{% ifversion actions-reusable-workflow-matrix %}, strategy, matrix{% endif %}{% ifversion actions-unified-inputs %}, inputs{% endif %}` | |
| `on.workflow_call.inputs.<inputs_id>.default` | `github{% ifversion actions-unified-inputs %}, inputs{% endif %}` | |
| `on.workflow_call.outputs.<output_id>.value` | `github, jobs, inputs` | |
{% else %}
| Path | Context | Special functions |
| ---- | ------- | ----------------- |
| `concurrency` | `github` | |
| `env` | `github, secrets` | |
| `jobs.<job_id>.concurrency` | `github, needs, strategy, matrix` | |
| `jobs.<job_id>.container` | `github, needs, strategy, matrix` | |
| `jobs.<job_id>.container.credentials` | `github, needs, strategy, matrix, env, secrets` | |
| `jobs.<job_id>.container.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets` | |
| `jobs.<job_id>.continue-on-error` | `github, needs, strategy, matrix` | |
| `jobs.<job_id>.defaults.run` | `github, needs, strategy, matrix, env` | |
| `jobs.<job_id>.env` | `github, needs, strategy, matrix, secrets` | |
| `jobs.<job_id>.environment` | `github, needs, strategy, matrix` | |
| `jobs.<job_id>.environment.url` | `github, needs, strategy, matrix, job, runner, env, steps` | |
| `jobs.<job_id>.if` | `github, needs` | `always, cancelled, success, failure` |
| `jobs.<job_id>.name` | `github, needs, strategy, matrix` | |
| `jobs.<job_id>.outputs.<output_id>` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | |
| `jobs.<job_id>.runs-on` | `github, needs, strategy, matrix` | |
| `jobs.<job_id>.services` | `github, needs, strategy, matrix` | |
| `jobs.<job_id>.services.<service_id>.credentials` | `github, needs, strategy, matrix, env, secrets` | |
| `jobs.<job_id>.services.<service_id>.env.<env_id>` | `github, needs, strategy, matrix, job, runner, env, secrets` | |
| `jobs.<job_id>.steps.continue-on-error` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.env` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.if` | `github, needs, strategy, matrix, job, runner, env, steps` | `always, cancelled, success, failure, hashFiles` |
| `jobs.<job_id>.steps.name` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.run` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.timeout-minutes` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.with` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.steps.working-directory` | `github, needs, strategy, matrix, job, runner, env, secrets, steps` | `hashFiles` |
| `jobs.<job_id>.strategy` | `github, needs` | |
| `jobs.<job_id>.timeout-minutes` | `github, needs, strategy, matrix` | |
{% endif %}
