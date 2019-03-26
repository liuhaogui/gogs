// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"github.com/gogs/git-module"

	"github.com/gogs/gogs/models"
	"github.com/gogs/gogs/pkg/context"
	"github.com/gogs/gogs/routes/repo"
	"fmt"
	"encoding/json"
	api "github.com/gogs/go-gogs-client"
)

// https://github.com/gogs/go-gogs-client/wiki/Repositories-Contents#download-raw-content
func GetRawFile(c *context.APIContext) {
	if !c.Repo.HasAccess() {
		c.Status(404)
		return
	}

	if c.Repo.Repository.IsBare {
		c.Status(404)
		return
	}

	blob, err := c.Repo.Commit.GetBlobByPath(c.Repo.TreePath)
	if err != nil {
		if git.IsErrNotExist(err) {
			c.Status(404)
		} else {
			c.Error(500, "GetBlobByPath", err)
		}
		return
	}
	if err = repo.ServeBlob(c.Context, blob); err != nil {
		c.Error(500, "ServeBlob", err)
	}
}

func GetRawFiles(c *context.APIContext) {
	var (
		fileList       []api.DiffFileList
		retrunFileList []api.ReturnDiffFile
	)
	requestBody := c.Query("fileList")
	if len(requestBody) < 1 {
		requestBody, _ = c.Context.Req.Body().String()
	}

	err := json.Unmarshal([]byte(requestBody), &fileList)
	if err != nil {
		fmt.Println(err)
		//todo
	}

	for _, v := range fileList {
		diffFileInfo := api.ReturnDiffFile{
			BaseInfo: v,
		}
		if v.IsBinary == false {
			repoPath := models.RepoPath(v.ProjectOwner, v.Project)
			//todo 容错判断
			c.Repo.GitRepo, err = git.OpenRepository(repoPath)
			if err != nil{
				fmt.Println("GetRawFiles OpenRepository err ",err)
			}
			c.Repo.TreePath = v.File

			c.Repo.CommitID = v.BaseDiffBranchCommitID
			c.Repo.Commit, err = c.Repo.GitRepo.GetCommit(v.BaseDiffBranchCommitID)
			blob, err := c.Repo.Commit.GetBlobByPath(c.Repo.TreePath)
			if err != nil{
				fmt.Println("GetBlobByPath err " ,err)
				diffFileInfo.BaseDiffFile = ""
			} else {
				r, err := blob.Blob().Data()
				if err != nil{
					fmt.Println("r, err := blob.Blob().Data() ",err)
				}
				diffFileInfo.BaseDiffFile = fmt.Sprintf("%s",r)
			}

			c.Repo.CommitID = v.DeployBranchCommitID
			c.Repo.Commit, err = c.Repo.GitRepo.GetCommit(v.DeployBranchCommitID)
			blob2, _ := c.Repo.Commit.GetBlobByPath(c.Repo.TreePath)
			r2, _ := blob2.Blob().Data()

			diffFileInfo.BranchDiffFile = fmt.Sprintf("%s",r2)
		} else {
			diffFileInfo.BaseDiffFile = "[gogs提示]二进制文件暂不提供显示。"
			diffFileInfo.BranchDiffFile = ""
		}
		retrunFileList = append(retrunFileList, diffFileInfo)
	}
	c.JSON(200, retrunFileList)
	return
}

// https://github.com/gogs/go-gogs-client/wiki/Repositories-Contents#download-archive
func GetArchive(c *context.APIContext) {
	repoPath := models.RepoPath(c.Params(":username"), c.Params(":reponame"))
	gitRepo, err := git.OpenRepository(repoPath)
	if err != nil {
		c.Error(500, "OpenRepository", err)
		return
	}
	c.Repo.GitRepo = gitRepo

	repo.Download(c.Context)
}

func GetEditorconfig(c *context.APIContext) {
	ec, err := c.Repo.GetEditorconfig()
	if err != nil {
		if git.IsErrNotExist(err) {
			c.Error(404, "GetEditorconfig", err)
		} else {
			c.Error(500, "GetEditorconfig", err)
		}
		return
	}

	fileName := c.Params("filename")
	def := ec.GetDefinitionForFilename(fileName)
	if def == nil {
		c.Error(404, "GetDefinitionForFilename", err)
		return
	}
	c.JSON(200, def)
}
