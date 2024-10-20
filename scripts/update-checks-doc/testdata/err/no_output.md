<a id="hello"></a>
## Hello

Example input:

```yaml
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo ${{ unknown }}
```

[Playground](https://rhysd.github.io/actionlint/#THIS_URL_WILL_BE_UPDATED)
