# ghrelease

A simple command line tool that creates a github release and adds specific files to it.

## Usage
1. Create a json file (let's call it `release.json`) describing your release. E.g.

    ```
    {
        "github_host": "https://github.com",
        "owner": "zerogvt",
        "repo": "ghrelease",
        "files": ["bin/ghrelease_lin", "bin/ghrelease_osx", "bin/ghrelease_win"],
        "tag": "latest",
        "desc": "description"
    }s
    ```
    `github_host` should always be `https://github.com`.
    
    `owner` is the owner (user or org) of the target repo.
    
    `repo` is the target repo where the release is to be hosted.
    
    `files` is a list of files to me uploaded. Paths should be relative to the path of `release.json`
    
    `tag` is the release tag
    
    `desc` is a description added to your release

2. [Download](https://github.ibm.com/vasigkou/tools/releases/tag/latest) and mark as executable (e.g. `chmod +x ghrelease_osx`) the correct version of the tool for your platform (ghrelease_lin for Linux, ghrelease_osx for OSx). Linux build covers Travis ubuntu workers and OSx covers Darwin platforms.
   Note: If you want to download within a pipeline (i.e. in an automated env) use the bash script `ghet.sh`. See the script for an example usage.

3. Have a github access token handy and export it in your env:
    `export GITHUB_TOKEN="your_githubb_token"

4. Run the tool giving it the path to your `release.json`: 
   
   `./ghrelease_osx -settings=release.json` if you're on OSx or
   
   `./ghrelease_lin -settings=release.json` if you're on a Linux platform. 

## Important. Do not shoot your foot!
If you try to create a release that already exists the tool will **replace** the existing release with the new one.

## Download links
To keep things simple you can always download the latest version of the tool using next links: (TODO upd these when/if project graduates)
1. OSx version: https://github.com/zerogvt/ghrelease/releases/download/latest/ghrelease_osx
2. Linux version: https://github.com/vasigkou/ghrelease/releases/download/latest/ghrelease_lin

## Build
`make`

## Run tests
`make test`
