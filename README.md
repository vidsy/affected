# affected
A Go tool that determines Go Packages that have been affected by changes to other packages between VCS commits.

# How tos

1. (Optional) Install the tool:
```
go install github.com/vidsy/affected/cmd/affected
```

The installed location will need to be in your `$PATH` for you to use the
binary. Alternatively, use the `go run` method to run the tool against your
target

2. How to use the tool to identify changed services

- Navigate to the folder you want to run the tool from
```
cd /Users/dhanakanesh/Vidsy/back-end
``` 
- Run the command:
```
go run github.com/vidsy/affected/cmd/affected \
  -a origin/master \
  -b HEAD \
  -f json \
  group --pkg-prefix github.com/vidsy/back-end/services \
  --after 1 \
  -i="/**/services/*/.service.yaml" \
  -i="/**/services/*/Makefile" \
  -i="/**/services/*/Dockerfile" \
  -i="/**/services/*/VERSION" \
  -i="/**/services/*/config/*.json"
```
**This command can silently fail. Ensure the Git branch it is run from is up to date with master if it does**

- Run affected, excluding a specific filepath:
```
go run github.com/vidsy/affected/cmd/affected -x="/**/services/project-rpc/*" -a origin/master -b HEAD -f json
```
go run github.com/vidsy/affected/cmd/affected \
  -a origin/master \
  -b HEAD \
  -f json \
  group --pkg-prefix github.com/vidsy/back-end/services \
  --after 1 \
  -i="/**/services/*/.service.yaml" \
  -i="/**/services/*/Makefile" \
  -i="/**/services/*/Dockerfile" \
  -i="/**/services/*/VERSION" \
  -i="/**/services/*/config/*.json"
  -x="/**/services/project-rpc/*" \
```

TODO: Document remaining options
