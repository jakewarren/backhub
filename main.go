package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Songmu/prompter"
	"github.com/google/go-github/github"
	"github.com/remeh/sizedwaitgroup"
	"golang.org/x/oauth2"
)

// App holds everything we need to operate on
type App struct {
	config  *Config
	client  *github.Client
	shallow bool
}

func main() {

	usr, err := user.Current()
	if err != nil {
		fmt.Println("can not get current user", err)
		os.Exit(1)
	}

	var (
		threads    int
		configPath string
	)

	app := App{}

	flag.BoolVar(&app.shallow, "s", false, "shallow clone/pull")
	flag.IntVar(&threads, "t", 10, "number of goroutines to spawn")
	flag.StringVar(&configPath, "config", usr.HomeDir+"/.gorespect.json", "Path to config file")
	flag.Parse()

	wd, _ := os.Getwd()

	var confirmCorrectDirectory bool = prompter.YN(fmt.Sprintf("⚠ Warning, this app writes to the current working directory (%s).\n\nAre you sure you want to continue?", wd), false)
	if !confirmCorrectDirectory {
		os.Exit(0)
	}

	cs := NewConfigStorage(configPath)
	app.config = cs.Load()
	defer cs.Save(app.config)

	err = setUpConfig(app.config, os.Stdout, os.Stdin)
	if err != nil {
		fmt.Printf("Could not set up config: %s\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: app.config.Github.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	app.client = github.NewClient(tc)

	swg := sizedwaitgroup.New(threads)

	// backup starred repos
	starredRepos := app.getGithubStars()
	for _, r := range starredRepos {
		swg.Add()
		go func(r string) {
			app.syncRepo("starred", r)
			defer swg.Done()
		}(r)
	}

	// backup user repos
	userRepos := app.getGithubRepos()
	for _, r := range userRepos {
		swg.Add()
		go func(r string) {
			app.syncRepo("repos", r)
			defer swg.Done()
		}(r)
	}

	// backup starred gists
	starredGists := app.getStarredGists()
	for _, g := range starredGists {
		app.syncGist("starred-gists", g)
	}

	// backup user gists
	userGists := app.getGithubGists()
	for _, g := range userGists {
		app.syncGist("gists", g)
	}
}

func (a *App) syncGist(key string, g *github.Gist) {
	repoDir := key + "/" + g.GetID()

	exitCode := 0
	_, err := os.Stat(filepath.Join("./", repoDir, "/.git"))
	if err == nil {

		log.Printf("repo '%s' already exists, attempting update.\n", repoDir)
		var c *exec.Cmd
		if a.shallow {
			c = command("git", "pull", "--depth=1")
		} else {
			c = command("git", "pull")
		}

		c.Dir = repoDir

		if err = c.Run(); err != nil {
			log.Printf("git pull error: %s\n", err)
		}
		return
	}

	if mkdirErr := os.MkdirAll(repoDir, os.ModePerm); mkdirErr != nil {
		log.Fatalln(mkdirErr)
	}

	url := g.GetGitPullURL()
	log.Printf("git clone %s into %s", url, repoDir)
	var c *exec.Cmd
	if a.shallow {
		c = command("git", "clone", "--depth=1", "--shallow-submodules", "--single-branch", url, repoDir)
	} else {
		c = command("git", "clone", url, repoDir)
	}
	if cloneErr := c.Run(); cloneErr != nil {
		if exiterr, ok := cloneErr.(*exec.ExitError); ok {
			if status, exitok := exiterr.Sys().(syscall.WaitStatus); exitok {
				exitCode = status.ExitStatus()
			}
		}
		log.Printf("close %s failed, exit code %d, err: %v\n", repoDir, exitCode, cloneErr)
		_, err = os.Stat(repoDir)
		if err != nil {
			os.RemoveAll(repoDir)
		}
	}
}

func (a *App) syncRepo(key, fullname string) {
	repoDir := key + "/" + fullname
	repoInfo := strings.Split(fullname, "/")
	repoUser := repoInfo[0]
	repoName := repoInfo[1]

	exitCode := 0
	_, err := os.Stat(filepath.Join("./", repoDir, "/.git"))
	if err == nil {

		log.Printf("repo '%s' already exists, attempting update.\n", repoDir)
		var c *exec.Cmd
		if a.shallow {
			c = command("git", "pull", "--depth=1")
		} else {
			c = command("git", "pull")
		}

		c.Dir = repoDir

		if err = c.Run(); err != nil {
			log.Printf("git pull error: %s\n", err)
		}
		return
	}

	if mkdirErr := os.MkdirAll(repoDir, os.ModePerm); mkdirErr != nil {
		log.Fatalln(mkdirErr)
	}

	url := fmt.Sprintf("https://github.com/%s/%s.git", repoUser, repoName)
	log.Printf("git clone %s into %s", url, repoDir)
	var c *exec.Cmd
	if a.shallow {
		c = command("git", "clone", "--depth=1", "--shallow-submodules", "--single-branch", url, repoDir)
	} else {
		c = command("git", "clone", url, repoDir)
	}
	if cloneErr := c.Run(); cloneErr != nil {
		if exiterr, ok := cloneErr.(*exec.ExitError); ok {
			if status, exitok := exiterr.Sys().(syscall.WaitStatus); exitok {
				exitCode = status.ExitStatus()
			}
		}
		log.Printf("close %s failed, exit code %d, err: %v\n", repoDir, exitCode, cloneErr)
		_, err = os.Stat(repoDir)
		if err != nil {
			os.RemoveAll(repoDir)
		}
	}
}

func setUpConfig(config *Config, out io.Writer, in io.Reader) error {
	var err error

	if config.Github.Username == "" {
		config.Github.Username, err = promptGithubUsername(out, in)
		if err != nil {
			return err
		}
	}

	if config.Github.Token == "" {
		config.Github.Token, err = promptGithubToken(out, in)
		if err != nil {
			return err
		}
	}

	return nil
}

func command(name string, args ...interface{}) *exec.Cmd {
	var a []string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			a = append(a, v)
		case []string:
			a = append(a, v...)
		}
	}
	c := exec.Command(name, a...)
	c.Stderr = os.Stderr
	// ensure that git doesn't prompt for credentials on repos that have been taken down with DMCA notices
	//   Reference: https://blog.github.com/2015-02-06-git-2-3-has-been-released/#the-credential-subsystem-is-now-friendlier-to-scripting
	c.Env = append(c.Env, "GIT_TERMINAL_PROMPT=0")

	return c
}
