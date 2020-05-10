package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"channels/storage"
)

var (
	errEventIgnored = errors.New("ignored")
)

type githubMessage struct {
	Ref        string
	RefType    string `json:"ref_type"`
	Repository struct {
		ID       int64
		Name     string
		FullName string `json:"full_name"`
		HtmlURL  string `json:"html_url"`
	}
	Sender struct {
		Login string
	}
	Action     string
	HeadCommit struct {
		Message string
		URL     string
	} `json:"head_commit"`
	Commits []struct {
		ID      string
		Message string
		Author  struct {
			Name  string
			Email string
		}
		URL      string
		Distinct bool
	}
	Issue struct {
		Title   string
		HtmlURL string `json:"html_url"`
		Number  int64
	}
	PullRequest struct {
		Title   string
		HtmlURL string `json:"html_url"`
		Number  int64
		Merged  bool
	} `json:"pull_request"`
}

// webhookGitHub handles request from github as a webhook
func (e *env) webhookGitHub(c *gin.Context) {
	caller, ok := e.checkToken(c)
	if !ok {
		return
	}
	if len(caller.Caps) != 1 {
		c.AbortWithStatusJSON(500, gin.H{"error": "caps invalid"})
		return
	}

	var msg githubMessage
	var err error

	err = c.BindJSON(&msg)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	m := storage.Message{
		From:      caller.Name,
		To:        caller.Caps[0],
		Timestamp: time.Now().UnixNano(),
	}

	event := c.GetHeader("X-GitHub-Event")
	switch event {
	case "delete":
		m.Text, m.Markdown, err = messageFromGithubDelete(&msg)
	case "push":
		m.Text, m.Markdown, err = messageFromGithubPush(&msg)
	case "issues":
		m.Text, m.Markdown, err = messageFromGithubIssues(&msg)
	case "pull_request":
		m.Text, m.Markdown, err = messageFromGithubPullRequest(&msg)
	default:
		err = errEventIgnored
	}
	if err != nil {
		c.JSON(200, gin.H{"status": err.Error()})
		return
	}

	err = e.store.Save(&m)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func messageFromGithubDelete(msg *githubMessage) (text string, markdown string, err error) {
	text = fmt.Sprintf("[%s] %s deleted %s:%s",
		msg.Repository.FullName,
		msg.Sender.Login, msg.RefType,
		msg.Ref,
	)
	markdown = fmt.Sprintf("<%s|[%s]> %s deleted %s:%s",
		msg.Repository.HtmlURL, msg.Repository.FullName,
		msg.Sender.Login, msg.RefType,
		msg.Ref,
	)
	return
}

func messageFromGithubPush(msg *githubMessage) (text string, markdown string, err error) {
	if len(msg.Commits) == 0 {
		err = errEventIgnored
		return
	}
	branch := strings.TrimPrefix(msg.Ref, "refs/heads/")

	text = fmt.Sprintf("[%s:%s] %s pushed commits:\n",
		msg.Repository.FullName, branch, msg.Sender.Login,
	)
	markdown = fmt.Sprintf("<%s|[%s:%s]> %s pushed commits:\n",
		msg.Repository.HtmlURL, msg.Repository.FullName,
		branch, msg.Sender.Login,
	)
	for _, commit := range msg.Commits {
		sha := commit.ID
		if len(sha) > 7 {
			sha = commit.ID[:6]
		}

		text += fmt.Sprintf("-> %s@%s{%s}\n",
			sha, commit.Author.Name,
			strings.SplitN(commit.Message, "\n", 2)[0])
		markdown += fmt.Sprintf("> `<%s|%s>` %s - %s",
			commit.URL, sha, commit.Message, commit.Author.Name)
	}
	return
}

func messageFromGithubPullRequest(msg *githubMessage) (text string, markdown string, err error) {
	text = fmt.Sprintf("[%s] %s %s issue #%d\n{%s}\n( %s )",
		msg.Repository.FullName,
		msg.Sender.Login, msg.Action,
		msg.Issue.Number, msg.Issue.Title, msg.Issue.HtmlURL,
	)
	markdown = fmt.Sprintf("<%s|[%s]> %s %s issue <%s|#%d %s>",
		msg.Repository.HtmlURL, msg.Repository.FullName,
		msg.Sender.Login, msg.Action,
		msg.Issue.HtmlURL, msg.Issue.Number, msg.Issue.Title,
	)
	return
}

func messageFromGithubIssues(msg *githubMessage) (text string, markdown string, err error) {
	if msg.Action == "synchronize" || msg.Action == "edited" {
		err = errEventIgnored
		return
	}
	if msg.Action == "closed" && msg.PullRequest.Merged {
		msg.Action = "merged"
	}
	text = fmt.Sprintf("[%s] %s %s pull request #%d\n{%s}\n( %s )",
		msg.Repository.FullName,
		msg.Sender.Login, msg.Action,
		msg.PullRequest.Number, msg.PullRequest.Title, msg.PullRequest.HtmlURL,
	)
	markdown = fmt.Sprintf("<%s|[%s]> %s %s pull request <%s|#%d %s>",
		msg.Repository.HtmlURL, msg.Repository.FullName,
		msg.Sender.Login, msg.Action, msg.PullRequest.HtmlURL,
		msg.PullRequest.Number, msg.PullRequest.Title,
	)
	return
}
