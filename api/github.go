package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"channels/storage"
)

type githubMessage struct {
	Repository struct {
		ID       int64
		Name     string
		FullName string `json:"full_name"`
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
	if err := c.BindJSON(&msg); err != nil {
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

	case "push":
		m.Text = fmt.Sprintf("[%s] %s pushed commits:\n",
			msg.Repository.FullName, msg.Sender.Login,
		)
		for _, commit := range msg.Commits {
			sha := commit.ID
			if len(sha) > 7 {
				sha = commit.ID[:6]
			}

			m.Text += fmt.Sprintf("> %s@%s{%s}\n",
				sha, commit.Author.Name,
				strings.SplitN(commit.Message, "\n", 2)[0])
		}

	case "issues":
		m.Text = fmt.Sprintf("[%s] %s %s issue #%d\n>{%s}\n( %s )",
			msg.Repository.FullName, msg.Sender.Login, msg.Action,
			msg.Issue.Number, msg.Issue.Title, msg.Issue.HtmlURL,
		)

	case "pull_request":
		if msg.Action == "synchronize" || msg.Action == "edited" {
			c.JSON(200, gin.H{"status": "ignored"})
			return
		}
		if msg.Action == "closed" && msg.PullRequest.Merged {
			msg.Action = "merged"
		}
		m.Text = fmt.Sprintf("[%s] %s %s pull request #%d\n>{%s}\n( %s )",
			msg.Repository.FullName, msg.Sender.Login, msg.Action,
			msg.PullRequest.Number, msg.PullRequest.Title, msg.PullRequest.HtmlURL,
		)

	default:
		c.JSON(200, gin.H{"status": "ignored"})
		return
	}

	if err := e.store.Save(&m); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
