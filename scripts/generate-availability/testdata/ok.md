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
| <code>concurrency</code> | <code>github, inputs</code> | None |
| <code>env</code> | <code>github, secrets, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.concurrency</code> | <code>github, needs, strategy, matrix, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.container</code> | <code>github, needs, strategy, matrix, env, secrets, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.container.credentials</code> | <code>github, needs, strategy, matrix, env, secrets, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.container.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.continue-on-error</code> | <code>github, needs, strategy, matrix, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.defaults.run</code> | <code>github, needs, strategy, matrix, env, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.env</code> | <code>github, needs, strategy, matrix, secrets, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.environment</code> | <code>github, needs, strategy, matrix, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.environment.url</code> | <code>github, needs, strategy, matrix, job, runner, env, steps, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.if</code> | <code>github, needs, inputs</code> | <code>always, cancelled, success, failure</code> |
| <code>jobs.&lt;job_id&gt;.name</code> | <code>github, needs, strategy, matrix, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.outputs.&lt;output_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.runs-on</code> | <code>github, needs, strategy, matrix, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.secrets.&lt;secrets_id&gt;</code> | <code>github, needs,{% ifversion actions-reusable-workflow-matrix %} strategy, matrix,{% endif %} secrets{% ifversion actions-unified-inputs %}, inputs{% endif %}</code> | None |
| <code>jobs.&lt;job_id&gt;.services</code> | <code>github, needs, strategy, matrix, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.credentials</code> | <code>github, needs, strategy, matrix, env, secrets, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.steps.continue-on-error</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.env</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.if</code> | <code>github, needs, strategy, matrix, job, runner, env, steps, inputs</code> | <code>always, cancelled, success, failure, hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.name</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.run</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.timeout-minutes</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.with</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.working-directory</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.strategy</code> | <code>github, needs, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.timeout-minutes</code> | <code>github, needs, strategy, matrix, inputs</code> | None |
| <code>jobs.&lt;job_id&gt;.with.&lt;with_id&gt;</code> | <code>github, needs{% ifversion actions-reusable-workflow-matrix %}, strategy, matrix{% endif %}{% ifversion actions-unified-inputs %}, inputs{% endif %}</code> | None |
| <code>on.workflow_call.inputs.&lt;inputs_id&gt;.default</code> | <code>github{% ifversion actions-unified-inputs %}, inputs{% endif %}</code> | None |
| <code>on.workflow_call.outputs.&lt;output_id&gt;.value</code> | <code>github, jobs, inputs</code> | None |
{% else %}
| Path | Context | Special functions |
| ---- | ------- | ----------------- |
| <code>concurrency</code> | <code>github</code> | None |
| <code>env</code> | <code>github, secrets</code> | None |
| <code>jobs.&lt;job_id&gt;.concurrency</code> | <code>github, needs, strategy, matrix</code> | None |
| <code>jobs.&lt;job_id&gt;.container</code> | <code>github, needs, strategy, matrix</code> | None |
| <code>jobs.&lt;job_id&gt;.container.credentials</code> | <code>github, needs, strategy, matrix, env, secrets</code> | None |
| <code>jobs.&lt;job_id&gt;.container.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets</code> | None |
| <code>jobs.&lt;job_id&gt;.continue-on-error</code> | <code>github, needs, strategy, matrix</code> | None |
| <code>jobs.&lt;job_id&gt;.defaults.run</code> | <code>github, needs, strategy, matrix, env</code> | None |
| <code>jobs.&lt;job_id&gt;.env</code> | <code>github, needs, strategy, matrix, secrets</code> | None |
| <code>jobs.&lt;job_id&gt;.environment</code> | <code>github, needs, strategy, matrix</code> | None |
| <code>jobs.&lt;job_id&gt;.environment.url</code> | <code>github, needs, strategy, matrix, job, runner, env, steps</code> | None |
| <code>jobs.&lt;job_id&gt;.if</code> | <code>github, needs</code> | <code>always, cancelled, success, failure</code> |
| <code>jobs.&lt;job_id&gt;.name</code> | <code>github, needs, strategy, matrix</code> | None |
| <code>jobs.&lt;job_id&gt;.outputs.&lt;output_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | None |
| <code>jobs.&lt;job_id&gt;.runs-on</code> | <code>github, needs, strategy, matrix</code> | None |
| <code>jobs.&lt;job_id&gt;.services</code> | <code>github, needs, strategy, matrix</code> | None |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.credentials</code> | <code>github, needs, strategy, matrix, env, secrets</code> | None |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets</code> | None |
| <code>jobs.&lt;job_id&gt;.steps.continue-on-error</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.env</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.if</code> | <code>github, needs, strategy, matrix, job, runner, env, steps</code> | <code>always, cancelled, success, failure, hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.name</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.run</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.timeout-minutes</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.with</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.working-directory</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.strategy</code> | <code>github, needs</code> | None |
| <code>jobs.&lt;job_id&gt;.timeout-minutes</code> | <code>github, needs, strategy, matrix</code> | None |
{% endif %}

### Context availability 2

Different contexts are available throughout a workflow run. For example, the `secrets` context may only be used at certain places within a job.

In addition, some functions may only be used in certain places. For example, the `hashFiles` function is not available everywhere.

The following table indicates where each context and special function can be used within a workflow. Unless listed below, a function can be used anywhere.

{% ifversion fpt or ghes > 3.3 or ghae-issue-4757 or ghec %}

| Workflow key | Context | Special functions |
| ---- | ------- | ----------------- |
| <code>concurrency</code> | <code>github, inputs</code> | |
| <code>env</code> | <code>github, secrets, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.concurrency</code> | <code>github, needs, strategy, matrix, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.container</code> | <code>github, needs, strategy, matrix, env, secrets, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.container.credentials</code> | <code>github, needs, strategy, matrix, env, secrets, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.container.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.continue-on-error</code> | <code>github, needs, strategy, matrix, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.defaults.run</code> | <code>github, needs, strategy, matrix, env, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.env</code> | <code>github, needs, strategy, matrix, secrets, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.environment</code> | <code>github, needs, strategy, matrix, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.environment.url</code> | <code>github, needs, strategy, matrix, job, runner, env, steps, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.if</code> | <code>github, needs, inputs</code> | <code>always, cancelled, success, failure</code> |
| <code>jobs.&lt;job_id&gt;.name</code> | <code>github, needs, strategy, matrix, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.outputs.&lt;output_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.runs-on</code> | <code>github, needs, strategy, matrix, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.secrets.&lt;secrets_id&gt;</code> | <code>github, needs,{% ifversion actions-reusable-workflow-matrix %} strategy, matrix,{% endif %} secrets{% ifversion actions-unified-inputs %}, inputs{% endif %}</code> | |
| <code>jobs.&lt;job_id&gt;.services</code> | <code>github, needs, strategy, matrix, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.credentials</code> | <code>github, needs, strategy, matrix, env, secrets, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.steps.continue-on-error</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.env</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.if</code> | <code>github, needs, strategy, matrix, job, runner, env, steps, inputs</code> | <code>always, cancelled, success, failure, hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.name</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.run</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.timeout-minutes</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.with</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.working-directory</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps, inputs</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.strategy</code> | <code>github, needs, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.timeout-minutes</code> | <code>github, needs, strategy, matrix, inputs</code> | |
| <code>jobs.&lt;job_id&gt;.with.&lt;with_id&gt;</code> | <code>github, needs{% ifversion actions-reusable-workflow-matrix %}, strategy, matrix{% endif %}{% ifversion actions-unified-inputs %}, inputs{% endif %}</code> | |
| <code>on.workflow_call.inputs.&lt;inputs_id&gt;.default</code> | <code>github{% ifversion actions-unified-inputs %}, inputs{% endif %}</code> | |
| <code>on.workflow_call.outputs.&lt;output_id&gt;.value</code> | <code>github, jobs, inputs</code> | |
{% else %}
| Path | Context | Special functions |
| ---- | ------- | ----------------- |
| <code>concurrency</code> | <code>github</code> | |
| <code>env</code> | <code>github, secrets</code> | |
| <code>jobs.&lt;job_id&gt;.concurrency</code> | <code>github, needs, strategy, matrix</code> | |
| <code>jobs.&lt;job_id&gt;.container</code> | <code>github, needs, strategy, matrix</code> | |
| <code>jobs.&lt;job_id&gt;.container.credentials</code> | <code>github, needs, strategy, matrix, env, secrets</code> | |
| <code>jobs.&lt;job_id&gt;.container.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets</code> | |
| <code>jobs.&lt;job_id&gt;.continue-on-error</code> | <code>github, needs, strategy, matrix</code> | |
| <code>jobs.&lt;job_id&gt;.defaults.run</code> | <code>github, needs, strategy, matrix, env</code> | |
| <code>jobs.&lt;job_id&gt;.env</code> | <code>github, needs, strategy, matrix, secrets</code> | |
| <code>jobs.&lt;job_id&gt;.environment</code> | <code>github, needs, strategy, matrix</code> | |
| <code>jobs.&lt;job_id&gt;.environment.url</code> | <code>github, needs, strategy, matrix, job, runner, env, steps</code> | |
| <code>jobs.&lt;job_id&gt;.if</code> | <code>github, needs</code> | <code>always, cancelled, success, failure</code> |
| <code>jobs.&lt;job_id&gt;.name</code> | <code>github, needs, strategy, matrix</code> | |
| <code>jobs.&lt;job_id&gt;.outputs.&lt;output_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | |
| <code>jobs.&lt;job_id&gt;.runs-on</code> | <code>github, needs, strategy, matrix</code> | |
| <code>jobs.&lt;job_id&gt;.services</code> | <code>github, needs, strategy, matrix</code> | |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.credentials</code> | <code>github, needs, strategy, matrix, env, secrets</code> | |
| <code>jobs.&lt;job_id&gt;.services.&lt;service_id&gt;.env.&lt;env_id&gt;</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets</code> | |
| <code>jobs.&lt;job_id&gt;.steps.continue-on-error</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.env</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.if</code> | <code>github, needs, strategy, matrix, job, runner, env, steps</code> | <code>always, cancelled, success, failure, hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.name</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.run</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.timeout-minutes</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.with</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.steps.working-directory</code> | <code>github, needs, strategy, matrix, job, runner, env, secrets, steps</code> | <code>hashFiles</code> |
| <code>jobs.&lt;job_id&gt;.strategy</code> | <code>github, needs</code> | |
| <code>jobs.&lt;job_id&gt;.timeout-minutes</code> | <code>github, needs, strategy, matrix</code> | |
{% endif %}
