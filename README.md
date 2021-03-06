# git-all

git-all is a tool written in go that will pull or clone every repo in an organization on github, providing your api key has access.

## install

```
go get -v github.com/jrieck1991/git-all
go install github.com/jrieck1991/git-all
```

## git-all --help
```
Usage of git-all:
  -a string
        git api key
  -c int
        set number of CPUs this program is allowed to use, int < 1 will use max CPUs available (default 1)
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
export GITHUB_ORG=org
export GIT_CPUS=1
```

## Usage Examples

#### pull/clone all repos from the my-repo organization to the repos directory using env vars or ~/.gitconfig
```
git-all
```

#### pull/clone all repos from the my-repo organization to the repos directory using flags
```
git-all -o my-repo -u jrieck1991 -a XXXXXX -r repos/
```
