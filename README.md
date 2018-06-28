# git-all

git-all is a tool written in go that will pull or clone every repo in an organization on github, providing your api key has access.

## install

```
Install go
# change to src directory within go path
cd ~/go/src/
git clone git@github.com:jrieck1991/git-all.git
cd git-all
go get
# mac osx
GOOS=darwin go install .
# linux
GOOS=linux go install .
```

## git-all --help
```
Usage of git-all:
  -a string
        git api key
  -o string
        set organziation to query
  -r string
        directory to download git repos into
  -u string
        github username
```

It's possible to use ~/.gitconfig to set your github api key and user
```
git config --global github.user jrieck1991
git config --global github.token XXXXXXXXX

# ~/.gitconfig
[github]
	user = jrieck1991
	token = XXXXXXXXXXXX
```

Or environment variables for api key, repo & user name
```
export GIT_USER=jrieck1991
export REPO_DIR=repos/
export GITHUB_API_KEY=XXXXXXX
```

## Usage Examples

#### pull/clone all repos from the my-repo organization to the repos directory using env vars or ~/.gitconfig
```
git-all -o my-repo -r repos/
```

#### pull/clone all repos from the my-repo organization to the repos directory using flags
```
git-all -o my-repo -u jrieck1991 -a XXXXXX -r repos/
```
