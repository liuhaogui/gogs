// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	api "github.com/gogs/go-gogs-client"

	"github.com/gogs/gogs/models"
	"github.com/gogs/gogs/pkg/context"
	"github.com/gogs/gogs/routes/api/v1/convert"
	"github.com/gogs/git-module"
)

// https://github.com/gogs/go-gogs-client/wiki/Repositories#get-branch
func GetBranch(c *context.APIContext) {
	branch, err := c.Repo.Repository.GetBranch(c.Params("*"))
	if err != nil {
		if models.IsErrBranchNotExist(err) {
			c.Error(404, "GetBranch", err)
		} else {
			c.Error(500, "GetBranch", err)
		}
		return
	}

	commit, err := branch.GetCommit()
	if err != nil {
		c.Error(500, "GetCommit", err)
		return
	}

	c.JSON(200, convert.ToBranch(branch, commit))
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories#list-branches
func ListBranches(c *context.APIContext) {
	branches, err := c.Repo.Repository.GetBranches()
	if err != nil {
		c.Error(500, "GetBranches", err)
		return
	}

	apiBranches := make([]*api.Branch, len(branches))
	for i := range branches {
		commit, err := branches[i].GetCommit()
		if err != nil {
			c.Error(500, "GetCommit", err)
			return
		}
		apiBranches[i] = convert.ToBranch(branches[i], commit)
	}

	c.JSON(200, &apiBranches)
}

func DiffBranch(c *context.Context)  {
	userName := c.Repo.Owner.Name
	repoName := c.Repo.Repository.Name
	c.Repo.Owner.Name = userName
	c.Repo.Repository.Name = repoName
	branch1 := c.Params(":branch1")
	branch2 := c.Params(":branch2")

	if c.Repo.GitRepo == nil {
		repoPath := models.RepoPath(c.Repo.Owner.Name, c.Repo.Repository.Name)
		var err error
		c.Repo.GitRepo, err = git.OpenRepository(repoPath)
		if err != nil {
			c.Handle(500, "RepoRef Invalid repo "+repoPath, err)
			return
		}
	}

	res, err := c.Repo.GitRepo.DiffBranch(branch1,branch2)
	if err != nil{
		c.JSON(200, res)
		return
	}
	c.JSON(200, res)
}

