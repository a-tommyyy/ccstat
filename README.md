# ccstat
Conventional Commit Statistic  
Aggregate total changed line count by specific dimension.

## examples
```bash
$ ccstat --group-by=scope --after=2022-01-01 --before=2022-12-31
scope   insertion   deletion
----------------------------
core    1135        323
ci      29          10
None    123456      54321

$ ccstat --group-by=type
type    insertion   deletion
----------------------------
feat    123456      54321
fix     1135        323
build   29          10
```

## Installation
```bash
brew install atomiyama/tap/ccstat
```


## Flags
### --group-by

| option | description |
| :----- | :---------- |
| scope  | Conventional Commit Scope |
| type   | **(NOT IMPLEMENTED)** Conventional Commit Type |
| author | **(NOT IMPLEMENTED)** |
| committer | **(NOT IMPLEMENTED)** |
| date | **(NOT IMPLEMENTED)** |


### --after/--before
Aggregate commits more recent than a specific date.