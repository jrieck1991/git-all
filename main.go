package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type gitCreds struct {
	gitUser, gitAPIKey, repoDir, org string
}

const (
	ctxTimeout      = 300 * time.Second // timeout for entire application
	osCmdTimeout    = 240 * time.Second // timeout per git cmd
	repoListTimeout = 30 * time.Second  // timeout to pull list of repos from github
)

func main() {
	// get user specific creds
	creds := initUser()
	// gather local repos
	localRepos := listRepoDir(creds)

	// set application wide timeout with context
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	// init github client
	client := initGithubClient(ctx, creds)

	// gather repos for given org
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100, // max allowed by github per page
		},
	}

	// send repos to this channel
	repoChan := make(chan []*github.Repository, 10000)
	go func() {
		for {
			// Paginate through to get all repos
			more, err := getRepoList(ctx, creds.org, client, opt, repoChan)
			if err != nil {
				log.Println(err)
			}
			// Close channel to signal done collecting git repo list
			if !more {
				close(repoChan)
				return
			}
		}
	}()

	// iterate over repos
Loop:
	for {
		select {
		// checks if background context is done
		case <-ctx.Done():
			log.Println(ctx.Err(), "exiting...")
			os.Exit(2)
		// get repos from chan if there are any
		case repos, open := <-repoChan:
			iterateRepos(ctx, localRepos, repos, creds)
			if !open {
				break Loop
			}
		}
	}

}

// gets a page of repos and shoves them into a channel, returns false when no more
func getRepoList(ctx context.Context, org string, client *github.Client, opt *github.RepositoryListByOrgOptions, repoChan chan<- []*github.Repository) (bool, error) {
	// set timeout for getting repo lists from github
	repoListCTX, cancel := context.WithTimeout(ctx, repoListTimeout)
	defer cancel()
	// list repos, paginated
	repos, response, err := client.Repositories.ListByOrg(repoListCTX, org, opt)
	if err != nil {
		if ctx.Err() != nil {
			return true, ctx.Err()
		}
		return true, err
	}
	repoChan <- repos
	// exit at last page
	if response.NextPage == 0 {
		return false, nil
	}
	opt.Page = response.NextPage

	// check if rate limited
	if _, ok := err.(*github.RateLimitError); ok {
		return false, errors.New("hit github API rate limit")
	}
	return getRepoList(ctx, org, client, opt, repoChan)
}

// init github client
func initGithubClient(ctx context.Context, c *gitCreds) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.gitAPIKey},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}

// init user vars
func initUser() *gitCreds {
	creds := &gitCreds{}

	gitUser := flag.String("u", "", "github username")
	repoDir := flag.String("r", "", "directory to download git repos into")
	gitAPIKey := flag.String("a", "", "git api key")
	org := flag.String("o", "", "set organziation to query")
	flag.Parse()

	// check for env vars then ~/.gitconfig if flags aren't set
	if *gitUser == "" {
		*gitUser = os.Getenv("GIT_USER")
		if *gitUser == "" {
			*gitUser = findGitConfigLine("user =")
		}
		if *gitUser == "" {
			log.Fatalln("must specifiy github username")
		}
	}
	if *repoDir == "" {
		*repoDir = os.Getenv("REPO_DIR")
		if *repoDir == "" {
			log.Fatalln("must specifiy local directory to store git repos")
		}
	}
	if *gitAPIKey == "" {
		*gitAPIKey = os.Getenv("GITHUB_API_KEY")
		if *gitAPIKey == "" {
			*gitAPIKey = findGitConfigLine("token =")
		}
		if *gitAPIKey == "" {
			log.Fatalln("must specifiy a valid github user api key")
		}
	}
	if *org == "" {
		*org = os.Getenv("GITHUB_ORG")
		if *org == "" {
			log.Fatalln("must specifiy github organization to query")
		}
	}

	// populate creds struct
	creds = &gitCreds{
		gitUser:   *gitUser,
		gitAPIKey: *gitAPIKey,
		repoDir:   *repoDir,
		org:       *org,
	}
	return creds
}

// Parse git config for user
// expects ~/.gitconfig
func findGitConfigLine(s string) string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		log.Fatalln("Could not find user's home directory")
	}
	f := fmt.Sprintf("%s/.gitconfig", homeDir)
	file, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(file), "\n")
	for _, l := range lines {
		if strings.Contains(l, s) {
			return parseGitConfigLine(l)
		}
	}
	return ""
}

func parseGitConfigLine(s string) string {
	c := strings.Split(s, "=")
	return c[1]
}

// Exists returns true if repo currently exists in specified directory
func exists(repoName string, localRepos []os.FileInfo) bool {
	for _, r := range localRepos {
		if r == nil {
			continue
		}
		if !r.IsDir() {
			continue
		}
		if repoName == r.Name() {
			return true
		}
	}
	return false
}

// listRepoDir
func listRepoDir(c *gitCreds) []os.FileInfo {
	localRepos, err := ioutil.ReadDir(c.repoDir)
	if err != nil {
		log.Fatalln(err)
	}
	return localRepos
}

// getRepo pulls or clones the given repoName
func getRepo(ctx context.Context, gitCmd string, c *gitCreds, r *github.Repository, wg *sync.WaitGroup) {
	defer wg.Done()
	// git pull
	if gitCmd == "pull" {
		repo := fmt.Sprintf("%s/%s", c.repoDir, *r.Name)
		if err := exec.CommandContext(ctx, "git", "-C", repo, "pull").Run(); err != nil {
			if ctx.Err() != nil {
				log.Println(ctx.Err(), *r.Name)
				return
			}
			log.Println(err, *r.Name)
			return
		}
		log.Printf("pull successful for %s", *r.Name)
		return
	}
	// git clone
	if gitCmd == "clone" {
		if err := exec.CommandContext(ctx, "git", "-C", c.repoDir, "clone", *r.SSHURL).Run(); err != nil {
			if ctx.Err() != nil {
				log.Println(ctx.Err(), *r.Name)
				return
			}
			log.Println(err, *r.Name)
			return
		}
		log.Printf("clone successful for %s", *r.Name)
		return
	}
	log.Println(errors.New("unknown git command, choose 'pull' or 'clone'"))
}

// Git pull or clone on given repos
func iterateRepos(ctx context.Context, localRepos []os.FileInfo, repos []*github.Repository, c *gitCreds) {
	var wg sync.WaitGroup
	for _, r := range repos {
		wg.Add(1)
		// init child context to be passed to each git command
		childCTX, cancel := context.WithTimeout(ctx, osCmdTimeout)
		defer cancel()
		select {
		case <-childCTX.Done():
			err := fmt.Errorf("context deadline exceeded for repo %s", *r.Name)
			log.Println(err)
			return
		default:
			// check if repo exists locally
			if exists(*r.Name, localRepos) {
				go getRepo(childCTX, "pull", c, r, &wg)
				continue
			}
			go getRepo(childCTX, "clone", c, r, &wg)
		}
	}
	wg.Wait()
}
