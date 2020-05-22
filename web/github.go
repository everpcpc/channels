package web

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"channels/auth"
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
	Organization struct {
		Login string
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
func (s *Server) webhookGitHub(c *gin.Context) {
	ctxCaller, exists := c.Get("caller")
	if !exists {
		c.AbortWithStatusJSON(403, gin.H{"error": "caller not found"})
		return
	}
	caller := ctxCaller.(*auth.Caller)

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
		Source:    storage.MessageSourceWebhook,
		From:      caller.Name,
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

	if caller.Caps[0] != "#" {
		m.To = caller.Caps[0]
	} else {
		if msg.Organization.Login == "" {
			c.JSON(400, gin.H{"status": "no target"})
			return
		}
		m.To = "#" + msg.Organization.Login
	}

	err = s.store.Save(&m)
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
	markdown = fmt.Sprintf("<%s|[%s]> %s deleted %s `%s`",
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

	text = fmt.Sprintf("[%s:%s] %s pushed commits: ",
		msg.Repository.FullName, branch, msg.Sender.Login,
	)
	markdown = fmt.Sprintf("<%s|[%s:%s]> %s pushed commits:",
		msg.Repository.HtmlURL, msg.Repository.FullName,
		branch, msg.Sender.Login,
	)
	for _, commit := range msg.Commits {
		sha := commit.ID
		if len(sha) > 7 {
			sha = commit.ID[:7]
		}

		commitMessage := strings.SplitN(commit.Message, "\n", 2)[0]
		text += fmt.Sprintf("-> %s@%s{%s}\n",
			sha, commit.Author.Name, commitMessage)
		markdown += fmt.Sprintf("\n> `<%s|%s>` %s - %s",
			commit.URL, sha, commitMessage, commit.Author.Name)
	}
	return
}

func messageFromGithubIssues(msg *githubMessage) (text string, markdown string, err error) {
	if msg.Action == "edited" {
		err = errEventIgnored
		return
	}
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

func messageFromGithubPullRequest(msg *githubMessage) (text string, markdown string, err error) {
	switch msg.Action {
	case "synchronize", "edited", "review_requested":
		err = errEventIgnored
		return
	case "closed":
		if msg.PullRequest.Merged {
			msg.Action = "merged"
		}
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
