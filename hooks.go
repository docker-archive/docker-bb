package main

import (
	"encoding/json"
	"fmt"
)

type PushHook struct {
	Sha        string   `json:"after"`
	BeforeSha  string   `json:"before"`
	Sender     *User    `json:"sender"`
	Ref        string   `json:"ref"`
	RefName    string   `json:"ref_name"`
	Repo       *Repo    `json:"repository"`
	HeadCommit Commit   `json:"head_commit"`
	Commits    []Commit `json:"commits"`
}

type Commit struct {
	Id        string   `json:"id"`
	Url       string   `json:"url"`
	Message   string   `json:"message"`
	Timestamp string   `json:timestamp`
	Author    *Author  `json:"author"`
	Added     []string `json:"added"`
}

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type User struct {
	ID         int64  `json:"id"`
	Login      string `json:"login"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Company    string `json:"company"`
	Location   string `json:"location"`
	Blog       string `json:"blog"`
	Avatar     string `json:"avatar_url"`
	GravatarId string `json:"gravatar_id"`
	Url        string `json:"html_url"`
}

// Owner represents the owner of a Github Repository.
type Owner struct {
	Type  string `json:"type"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

// Permissions
type Permissions struct {
	Push  bool `json:"push"`
	Pull  bool `json:"pull"`
	Admin bool `json:"admin"`
}

type Source struct {
	Owner *Owner `json:"owner"`
}

// Repo represents a Github-hosted Git Repository.
type Repo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Fork     bool   `json:"fork"`
	SshUrl   string `json:"ssh_url"`
	GitUrl   string `json:"git_url"`
	CloneUrl string `json:"clone_url"`
	HtmlUrl  string `json:"html_url"`

	Owner       *Owner       `json:"owner"`
	Permissions *Permissions `json:"permissions"`
	Source      *Source      `json:"source"`
}

func ParseHook(raw []byte) (*PushHook, error) {
	hook := PushHook{}
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, err
	}

	// it is possible the JSON was parsed, however,
	// So we'll check to be sure certain key fields
	// were populated
	if hook.Ref != "refs/heads/master" {
		return nil, fmt.Errorf("This hook is not a commit to master, ref is %q", hook.Ref)
	}

	return &hook, nil
}
