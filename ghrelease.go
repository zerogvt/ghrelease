package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	gh "github.com/google/go-github/v32/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Release captures all release metadata we get from user settings
type Release struct {
	GithubHost string   `json:"github_host"`
	Owner      string   `json:"owner"`
	Repo       string   `json:"repo"`
	Files      []string `json:"files"`
	Tag        string   `json:"tag"`
	Desc       string   `json:"desc"`
}

// read in settings file into a Release struct
func readReleaseMeta(fpath string) Release {
	data, err := ioutil.ReadFile(fpath)
	quitOn(err)
	var rel Release
	err = json.Unmarshal(data, &rel)
	quitOn(err)
	return rel
}

// read env var GITHUB_TOKEN
func githubToken() string {
	GithubTokenEnvVar := "GITHUB_TOKEN"
	GithubToken := os.Getenv(GithubTokenEnvVar)
	if GithubToken == "" {
		log.Fatal(fmt.Sprintf("[ERROR] No env var %s\n", GithubTokenEnvVar))
	}
	return GithubToken
}

// quit if err
func quitOn(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// create a Github client
func ghClient(ctx context.Context, usr Release) *gh.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: githubToken(),
			TokenType:   "Bearer",
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	var client *gh.Client
	if usr.GithubHost == "https://github.com" {
		client = gh.NewClient(tc)
	} else {
		var err error
		client, err = gh.NewEnterpriseClient(
			usr.GithubHost+"/api/v3/",
			usr.GithubHost+"/api/uploads/", tc)
		quitOn(err)
	}
	return client
}

// parse command line args
func parseCLI(arguments []string) (string, bool) {
	var settings string
	var yolo bool
	flag.StringVar(&settings, "settings",
		"release.json", "path to settings file for this release")
	flag.BoolVar(&yolo, "yolo",
		false, "when yolo is set sanity checks are skipped")
	flag.CommandLine.Parse(arguments)
	return settings, yolo
}

type ghRepoSvc interface {
	GetReleaseByTag(
		ctx context.Context,
		owner, repo, tag string) (*gh.RepositoryRelease, *gh.Response, error)
	DeleteRelease(
		ctx context.Context, owner,
		repo string, id int64) (*gh.Response, error)
	CreateRelease(
		ctx context.Context, owner,
		repo string,
		release *gh.RepositoryRelease) (*gh.RepositoryRelease, *gh.Response, error)
	UploadReleaseAsset(
		ctx context.Context, owner string,
		repo string, id int64, opts *gh.UploadOptions,
		file *os.File) (*gh.ReleaseAsset, *gh.Response, error)
}

// Create a release.
// Deletes an existing one and replaces it with a new one.
func createRel(ctx context.Context,
	usr Release, reposvc ghRepoSvc) *gh.RepositoryRelease {
	log.Info("Creating release '" + usr.Tag + "'")
	rel, resp, err := reposvc.GetReleaseByTag(ctx,
		usr.Owner, usr.Repo, usr.Tag)
	// if release already exists delete it
	if resp.StatusCode != 404 {
		quitOn(err)
		_, err = reposvc.DeleteRelease(
			ctx, usr.Owner, usr.Repo, *rel.ID)
		quitOn(err)
	}
	fmt.Println(usr.Tag)
	rel, _, err = reposvc.CreateRelease(
		ctx, usr.Owner, usr.Repo,
		&gh.RepositoryRelease{
			TagName:         gh.String("usr.Tag"),
			TargetCommitish: gh.String("main"),
			Name:            gh.String(usr.Tag),
			Body:            gh.String(usr.Desc),
			Draft:           gh.Bool(false),
			Prerelease:      gh.Bool(false),
		})
	quitOn(err)
	return rel
}

// upload files to a specific release.
// Release must exist prior to that step.
func uploadFilesToRel(ctx context.Context, usr Release,
	rel *gh.RepositoryRelease, reposvc ghRepoSvc) {
	for _, fpath := range usr.Files {
		file, err := os.Open(fpath)
		quitOn(err)
		finfo, err := file.Stat()
		quitOn(err)
		log.Info("Uploading file " + finfo.Name() + " to release '" + *rel.TagName + "'")
		_, _, err = reposvc.UploadReleaseAsset(
			ctx, usr.Owner, usr.Repo,
			*rel.ID,
			&gh.UploadOptions{
				Name:  finfo.Name(),
				Label: finfo.Name(),
			},
			file)
		quitOn(err)
	}
}

// sanity checks whether things are looking sane
func sanity(usr *Release) {
	mydir, err := os.Getwd()
	quitOn(err)
	pathlist := strings.Split(mydir, string(os.PathSeparator))
	fmt.Println(pathlist)
	haveRepo := pathlist[len(pathlist)-1]
	wantRepo := usr.Repo
	if wantRepo != haveRepo {
		log.Fatal("You are in repo '" + haveRepo + "'" +
			" but trying to release to repo '" + wantRepo + "'")
	}
}

// set up logging module
func setupLogging() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}

// main
func main() {
	setupLogging()
	settingsfile, yolo := parseCLI(os.Args[1:])
	usr := readReleaseMeta(settingsfile)
	if !yolo {
		sanity(&usr)
	} else {
		log.Warn("I see you set yolo. Hope you know what you're doing.")
	}
	log.Info("Release setup: " + gh.Stringify(usr))
	ctx := context.Background()
	reposvc := ghClient(ctx, usr).Repositories
	rel := createRel(ctx, usr, reposvc)
	uploadFilesToRel(ctx, usr, rel, reposvc)
}
