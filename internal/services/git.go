package services

import (
	"fmt"
	"os"
	"strings"

	// "os/user"
	// "time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	// "github.com/go-git/go-git/v5/plumbing/object"
	//"github.com/go-git/go-git/v5/plumbing/transport"

	// Config "dotcomfy/internal/config"
	Log "dotcomfy/internal/logger"
)

func Clone(url, branch, commit_hash, path string) error {
	LOGGER = Log.GetLogger()
	// @REF [Basic go-git example](https://github.com/go-git/go-git/blob/master/_examples/clone/main.go)
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:               url, // Guaranteed at least one because cobra
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		ReferenceName:     plumbing.ReferenceName(branch),
		SingleBranch:      true,
	})
	if err != nil {
		LOGGER.Error(err)
		return err
	}

	if commit_hash != "" {
		worktree, err := repo.Worktree()
		if err != nil {
			LOGGER.Error(err)
			return err
		}

		err = worktree.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(commit_hash),
			Force: true,
		})
		if err != nil {
			LOGGER.Error(err)
			return err
		}
	}

	// head, err := repo.Head()
	// if err != nil {
	// 	LOGGER.Error(err)
	// 	return err
	// }

	// fmt.Println(head.Hash())

	return nil
}

func Pull(repo_path string) error {
	LOGGER = Log.GetLogger()
	repo, err := git.PlainOpen(repo_path)
	if err != nil {
		LOGGER.Errorf("Error opening the local repo in %s: %v", repo_path, err)
		return err
	}

	head, err := repo.Head()
	if err != nil {
		LOGGER.Errorf("Error getting HEAD: %v", err)
		return err
	}

	branch_name := string(head.Name())
	if strings.HasPrefix(branch_name, "refs/heads/") {
		branch_name = strings.TrimPrefix(branch_name, "refs/heads/")
	}

	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/" + branch_name + ":refs/remotes/origin/" + branch_name),
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		LOGGER.Errorf("Error fetching: %v", err)
		return err
	}

	remote_ref_name := plumbing.NewRemoteReferenceName("origin", branch_name)
	origin_ref, err := repo.Reference(remote_ref_name, true)
	if err != nil {
		LOGGER.Errorf("Error getting origin reference: %v", err)
		return err
	}

	LOGGER.Errorf("Origin ref after fetch: %s", origin_ref.Hash())

	branch := plumbing.NewBranchReferenceName(branch_name)
	// Bypass dirty worktree checks and just "fast forward" to the latest commit
	err = repo.Storer.SetReference(plumbing.NewHashReference(branch, origin_ref.Hash()))
	if err != nil {
		LOGGER.Errorf("Error switching local reference to latest from origin: %v", err)
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		LOGGER.Errorf("Error getting the worktree: %v", err)
		return err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branch,
		Force:  true,
	})
	if err != nil {
		LOGGER.Errorf("Error checking out branch: %v", err)
		return err
	}

	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: branch,
		SingleBranch:  true,
		Force:         true,
		Progress:      os.Stdout, // May omit this, we'll see how it looks
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		LOGGER.Errorf("Error pulling: %v", err)
		return err
	}

	head, err = repo.Head()
	if err != nil {
		LOGGER.Errorf("Error getting HEAD: %v", err)
		return err
	}

	fmt.Printf("HEAD is now at %s\n", head.Hash())

	return nil
}

/*
func Push(repo_path string) error {
	var repo_url string
	var branch string
	var failed_files []string

	config := Config.GetConfig()
	auth := config.Auth

	repo, err := git.PlainOpen(repo_path)
	if err != nil {
		LOGGER.Errorf("Error opening the local repo in %s: %v", repo_path, err)
		return err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		LOGGER.Error("Error getting remote from the origin:", err)
		return err
	}

	urls := remote.Config().URLs
	if len(urls) > 0 {
		repo_url = urls[0]
	} else {
		LOGGER.Fatal("No URL found for the remote 'origin'")
	}

	worktree, err := repo.Worktree()
	if err != nil {
		LOGGER.Errorf("Error getting the worktree: %v", err)
		return err
	}

	head, err := repo.Head()
	if err != nil {
		LOGGER.Errorf("Error getting HEAD: %v", err)
		return err
	}

	branch = head.Name().Short()

	status, err := worktree.Status()
	if err != nil {
		LOGGER.Errorf("Error getting status: %v", err)
		return err
	}

	if status.IsClean() {
		fmt.Println("No changes to commit")
		return nil
	}

	for file, s := range status {
		fmt.Println("Staging %s: %s", file, s.Worktree)
		_, err = worktree.Add(file)
		if err != nil {
			LOGGER.Errorf("Error adding %s: %v", file, err)
			failed_files = append(failed_files, file)
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	commit_message := fmt.Sprintf("Update from %s@%s at %s",
		username,
		hostname,
		time.Now().Format("12:30:00 CST 1963-11-22"),
	)

	_, err = worktree.Commit(commit_message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  auth.Username,
			Email: auth.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		LOGGER.Fatalf("Error committing: %v", err)
	}

	// TODO:
	// Set up auth for private repos and/or username/password auth at runtime
	// var auth *http.BasicAuth
	// if auth != nil {

	// }

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		Force:      false,
	})
	if err != nil {
		LOGGER.Fatalf("Error pushing: %v", err)
	}

	return nil
}
*/
