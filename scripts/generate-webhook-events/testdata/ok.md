---
title: Test
---

Wow

## Some header

Foo

## `workflow_dispatch`

| Webhook event payload | Activity types | `GITHUB_SHA` | `GITHUB_REF` |
| ------------------ | ------------ | ------------ | ------------------|
| [workflow_dispatch](/webhooks/event-payloads/#workflow_dispatch) | n/a | Last commit on the `GITHUB_REF` branch | Branch that received dispatch |

## About events that trigger workflows

Bar

## `check_run`

| Operator | Description | Example |
| -------- | ----------- | ------- |
| * | Any value | `* * * * *` runs every minute of every day. |

Aaaa aaaa

| Webhook event payload | Activity types | `GITHUB_SHA` | `GITHUB_REF` |
| --------------------- | -------------- | ------------ | -------------|
| [`check_run`](/webhooks/event-payloads/#check_run) | - `created`<br/>- `rerequested`<br/>- `completed` | Last commit on default branch | Default branch |

## `discussion`

Blah blah blah

| Webhook event payload | Activity types | `GITHUB_SHA` | `GITHUB_REF` |
| --------------------- | -------------- | ------------ | -------------|
| [`discussion`](/webhooks/event-payloads/#discussion) | - `opened`<br/>- `edited`<br/>- `deleted`<br/>- `transferred`<br/>- `pinned`<br/>- `unpinned`<br/>- `labeled`<br/>- `unlabeled`<br/>- `locked`<br/>- `unlocked`<br/>- `category_changed`<br/> - `answered`<br/> - `unanswered` | Last commit on default branch | Default branch |

## `create`

Runs your workflow

| Webhook event payload | Activity types | `GITHUB_SHA` | `GITHUB_REF` |
| --------------------- | -------------- | ------------ | -------------|
| [`create`](/webhooks/event-payloads/#create) | n/a | Last commit on the created branch or tag | Branch or tag created |

Table without row:

| Webhook event payload | Activity types | `GITHUB_SHA` | `GITHUB_REF` |
| --------------------- | -------------- | ------------ | -------------|
