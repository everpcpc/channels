package web

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"channels/auth"
	"channels/storage"
)

// alertManagerMessage from:
// https://prometheus.io/docs/alerting/configuration/#webhook_config
// https://github.com/prometheus/alertmanager/blob/66a0ed21bdb0720b4ba083d35acd6ae77fa7b0b5/template/template.go#L227
type alertManagerMessage struct {
	Version           string
	GroupKey          string
	Status            string
	Receiver          string
	GroupLabels       map[string]string
	CommonLabels      map[string]string
	CommonAnnotations map[string]string
	ExternalURL       string
	Alerts            []struct {
		Status       string
		Labels       map[string]string
		Annotations  map[string]string
		StartsAt     time.Time
		EndsAt       time.Time
		GeneratorURL string
		Fingerprint  string
	}
}

// webhookAlertManager handles request from alertmanager as a webhook
func (s *Server) webhookAlertManager(c *gin.Context) {
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
	if len(caller.Caps) != 1 {
		c.AbortWithStatusJSON(500, gin.H{"error": "caps invalid"})
		return
	}

	var msg alertManagerMessage
	if err := c.BindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	text := fmt.Sprintf("%s [%s:%s] is %s: %s\n( %s )\n",
		getStatusEmoji(msg.Status),
		msg.GroupLabels["alertname"], msg.Version,
		msg.CommonLabels["severity"],
		msg.CommonAnnotations["summary"],
		msg.ExternalURL,
	)
	markdown := fmt.Sprintf("%s <%s|[%s:%s] is %s: %s>\n",
		getStatusEmoji(msg.Status),
		msg.ExternalURL,
		msg.GroupLabels["alertname"], msg.Version,
		msg.CommonLabels["severity"],
		msg.CommonAnnotations["summary"],
	)
	var labels []string
	for k, v := range msg.CommonLabels {
		if k == "alertname" || k == "severity" {
			continue
		}
		labels = append(labels, k+"="+v)
	}
	text += "labels{" + strings.Join(labels, ",") + "}"
	markdown += "`labels{" + strings.Join(labels, ",") + "}`\n"

	for _, alert := range msg.Alerts {
		markdown += fmt.Sprintf("> <%s|%s>",
			alert.GeneratorURL, alert.Annotations["summary"])
	}
	m := storage.Message{
		Source:    storage.MessageSourceWebhook,
		From:      caller.Name,
		To:        caller.Caps[0],
		Text:      text,
		Markdown:  markdown,
		Timestamp: time.Now().UnixNano(),
	}

	if err := s.store.Save(&m); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func getStatusEmoji(status string) string {
	switch status {
	case "firing":
		return "🔥"
	case "resolved":
		return "✅"
	}
	return status
}
