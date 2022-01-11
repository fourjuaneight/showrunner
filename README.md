# Showrunner

A simple utility to get TV Show data from TMDB and use episode titles to rename media files (based on a specific naming convention) and add metadata.

This is mostly for personal use, but PRs are welcomed.

## Usage
There are 3 ways to run the script:

### [Gorun](https://github.com/erning/gorun#how-to-build-and-install-gorun-from-source)
```sh
make run
# script should run from root of repo
./showrunner.go
```

### Local Binary
```sh
make build
# binary should be accessible from the root of the repo
./showrunner
```

### [GOPATH Binary](https://github.com/fourjuaneight/dotfiles/blob/master/homedir/.zshenv#L16-L20)
```sh
make install
# binary should be accessible from anywhere
showrunner
```
