How to publish a new release?

Word about versions:
Releases are tag based, and follow (semantic versioning semantics)[semver.org]. e.g tag v0.3.3 is 
release 0.3.3 and under version 1.0 breaking changes occur.
The process is based on `git release` from git-extras package.

How to:
Fetch latest origin including tags
```console
git fetch --tags origin
```

Create a fresh branch to work on
```console
git checkout -b release origin/master
```

Create a 'minor' release, entering interactive changelog update:
```console
git release --semver minor -c 
```

This will automatically bump 0.5.0 to 0.6.0 for example. Use 'major', 'minor', 'patch'
where needed.

Now we have a new tagged commit, all we need it to push it:

```console
git push --tags origin/master
```

