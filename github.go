package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"log"

	"github.com/google/go-github/github"
)

const githubHost = "github.com"

var (
	// ErrCanNotGetUsername shows that we can not get username of github user
	ErrCanNotGetUsername = errors.New("can not get username")
	// ErrCanNotGetToken show that we can not get access token of github user
	ErrCanNotGetToken = errors.New("can not get token")
)

func promptGithubUsername(out io.Writer, in io.Reader) (string, error) {
	var username string
	_, err := prompt("Enter github username: ", &username, out, in)
	if err != nil {
		return "", ErrCanNotGetUsername
	}

	return username, nil
}

func promptGithubToken(out io.Writer, in io.Reader) (string, error) {
	var token string
	tokenURL := fmt.Sprintf("https://%s/settings/tokens/new?scopes=repo,gist&description=backhub", githubHost)
	message := fmt.Sprintf("Please, generate and copy token here: %s\nEnter token: ", tokenURL)

	_, err := prompt(message, &token, out, in)
	if err != nil {
		return "", ErrCanNotGetToken
	}

	return token, nil
}

func prompt(message string, dest interface{}, out io.Writer, in io.Reader) (int, error) {
	fmt.Fprint(out, message)

	return fmt.Fscan(in, dest)
}

// get all starred repositories for the authenticated user
func (a *App) getGithubStars() []string {

	var starredRepos []string
	opt := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: 500},
	}
	ctx := context.Background()
	for {
		repos, resp, err := a.client.Activity.ListStarred(ctx, a.config.Github.Username, opt)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range repos {
			starredRepos = append(starredRepos, *r.Repository.FullName)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return starredRepos
}

// get all repositories for the authenticated user
func (a *App) getGithubRepos() []string {

	var userRepos []string
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	ctx := context.Background()
	for {
		repos, resp, err := a.client.Repositories.List(ctx, a.config.Github.Username, opt)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range repos {
			userRepos = append(userRepos, *r.FullName)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return userRepos
}

// get all Gists for the authenticated user
func (a *App) getGithubGists() []*github.Gist {

	var userGists []*github.Gist
	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	ctx := context.Background()
	for {
		repos, resp, err := a.client.Gists.List(ctx, a.config.Github.Username, opt)
		if err != nil {
			log.Fatal(err)
		}

		userGists = append(userGists, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return userGists
}

// get all Gists for the authenticated user
func (a *App) getStarredGists() []*github.Gist {

	var starredGists []*github.Gist
	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	ctx := context.Background()
	for {
		repos, resp, err := a.client.Gists.ListStarred(ctx, opt)
		if err != nil {
			log.Fatal(err)
		}

		starredGists = append(starredGists, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return starredGists
}
