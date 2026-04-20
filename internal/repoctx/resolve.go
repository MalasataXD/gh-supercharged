package repoctx

import "github.com/cli/go-gh/v2/pkg/repository"

// CurrentRepo returns "owner/name" when the working directory is inside a
// github.com git repo, or "" otherwise. Never errors — no repo is normal.
func CurrentRepo() string {
	repo, err := repository.Current()
	if err != nil || repo.Host != "github.com" {
		return ""
	}
	return repo.Owner + "/" + repo.Name
}
