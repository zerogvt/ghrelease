package main

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"

	gh "github.com/google/go-github/v32/github"
)

func TestReadReleaseMeta(t *testing.T) {
	have := readReleaseMeta("testdata/release.json")
	want := Release{
		GithubHost: "https://github.com",
		Owner:      "zerogvt",
		Repo:       "ghrelease",
		Files:      []string{"bin/ghrelease_lin", "bin/ghrelease_osx"},
		Tag:        "latest",
		Desc:       "description",
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("Want %s but have %s", gh.Stringify(want), gh.Stringify(have))
	}
}

func TestGithubToken(t *testing.T) {
	want := "abc"
	os.Setenv("GITHUB_TOKEN", want)
	have := githubToken()
	if want != have {
		t.Errorf("Want %s but have %s", want, have)
	}
}

func TestQuitOn(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		err := errors.New("error msg")
		quitOn(err)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestQuitOn")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Errorf("process ran with err %v, want exit status 1", err)
}

func TestGHClient(t *testing.T) {
	ctx := context.Background()
	usr := Release{
		GithubHost: "https://github.com",
		Owner:      "zerogvt",
		Repo:       "ghrelease",
		Files:      []string{"bin/ghrelease_lin", "bin/ghrelease_osx"},
		Tag:        "latest",
		Desc:       "description",
	}
	client := ghClient(ctx, usr)
	wantBase, _ := url.Parse(usr.GithubHost + "/api/v3/")
	wantUpload, _ := url.Parse(usr.GithubHost + "/api/uploads/")
	if client.BaseURL.String() != wantBase.String() {
		t.Errorf("Bad client base url: \n" +
			client.BaseURL.String() + "\n" +
			wantBase.String())
	}
	if client.UploadURL.String() != wantUpload.String() {
		t.Errorf("Bad client upload urls: \n" +
			client.UploadURL.String() + "\n" +
			wantUpload.String())
	}
}

func TestParseCLIWithArgs(t *testing.T) {
	arguments := [...]string{"-yolo", "-settings", "another.json"}
	settingsfile, yolo := parseCLI(arguments[:])
	if !yolo {
		t.Errorf("yolo should be set")
	}
	if settingsfile != "another.json" {
		t.Errorf("settings file not set")
	}
}

const TestID = 123

func TestCreateRelease(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockghRepoSvc(ctrl)
	id := int64(TestID)
	tn := "test"
	rel := gh.RepositoryRelease{ID: &id, TagName: &tn}
	resp := gh.Response{Response: &http.Response{StatusCode: 200}}
	m.EXPECT().
		GetReleaseByTag(
			gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Return(&rel, &resp, nil)
	m.EXPECT().DeleteRelease(
		gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Times(1).Return(nil, nil)
	id2 := int64(123)
	rel2 := gh.RepositoryRelease{ID: &id2}
	m.EXPECT().CreateRelease(
		gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).Return(&rel2, nil, nil)

	res := createRel(
		context.Background(),
		Release{
			GithubHost: "https://github.com",
			Owner:      "zerogvt",
			Repo:       "ghrelease",
			Files:      []string{"bin/ghrelease_lin", "bin/ghrelease_osx"},
			Tag:        "latest",
			Desc:       "description",
		},
		m)
	if *res.ID != TestID {
		t.Errorf("Release ID is wrong")
	}
}

func TestUploadFilesToRel(t *testing.T) {
	tmpFile1, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	defer os.Remove(tmpFile1.Name())

	tmpFile2, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	defer os.Remove(tmpFile2.Name())

	ctrl := gomock.NewController(t)

	m := NewMockghRepoSvc(ctrl)
	ctx := context.Background()
	usr := Release{
		GithubHost: "https://github.com",
		Owner:      "zerogvt",
		Repo:       "ghrelease",
		Files:      []string{tmpFile1.Name(), tmpFile2.Name()},
		Tag:        "latest",
		Desc:       "description",
	}
	id := int64(TestID)
	tn := "test"
	rel := gh.RepositoryRelease{ID: &id, TagName: &tn}
	defer ctrl.Finish()
	m.EXPECT().
		UploadReleaseAsset(
			gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any()).Times(2)
	uploadFilesToRel(ctx, usr, &rel, m)
}
