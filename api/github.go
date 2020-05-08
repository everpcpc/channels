package api

import (
	"fmt"
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
	Issue struct {
		Title   string
		HtmlURL string `json:"html_url"`
		Number  int64
	}
	PullRequest struct {
		Title   string
		HtmlURL string `json:"html_url"`
		Number  int64
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
	msgToSend := &storage.Message{
		From:      caller.Name,
		To:        caller.Caps[0],
		Timestamp: time.Now().UnixNano(),
	}

	event := c.GetHeader("X-GitHub-Event")
	switch event {

	case "push":
		// TODO: show more commits
		msgToSend.Text = fmt.Sprintf("[%s] %s pushed commit ' %s ' ( %s )",
			msg.Repository.FullName, msg.Sender.Login,
			msg.HeadCommit.Message, msg.HeadCommit.URL,
		)

	case "issues":
		msgToSend.Text = fmt.Sprintf("[%s] %s %s issue #%d ' %s ' ( %s )",
			msg.Repository.FullName, msg.Sender.Login, msg.Action,
			msg.Issue.Number, msg.Issue.Title, msg.Issue.HtmlURL,
		)

	case "pull_request":
		msgToSend.Text = fmt.Sprintf("[%s] %s %s pull request #%d ' %s ' ( %s )",
			msg.Repository.FullName, msg.Sender.Login, msg.Action,
			msg.PullRequest.Number, msg.PullRequest.Title, msg.PullRequest.HtmlURL,
		)

	default:
		c.JSON(200, gin.H{"status": "ignored"})
		return
	}

	if err := e.store.Save(msgToSend); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
