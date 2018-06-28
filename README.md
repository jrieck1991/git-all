# git-all

# install

```
git clone git@github.com:jrieck1991/git-all.git
cd git-all
go get
GOOS=darwin go install .
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
# ~/.gitconfig
[github]
	user = jackrieck
	token = XXXXXXXXXXXX
```

Or environment variables for api key, repo & user name
```
export GIT_USER=jrieck
export REPO_DIR=repos/
export GITHUB_API_KEY=XXXXXXX
```
