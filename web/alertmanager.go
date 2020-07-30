package web

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	promtemplate "github.com/prometheus/alertmanager/template"
	"github.com/sirupsen/logrus"

	"channels/auth"
	"channels/storage"
)

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

	var msg promtemplate.Data
	if err := c.BindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	text := fmt.Sprintf("%s [%s] is %s: %s\n( %s )\n",
		getStatusEmoji(msg.Status),
		msg.GroupLabels["alertname"],
		msg.CommonLabels["severity"],
		msg.CommonAnnotations["summary"],
		msg.ExternalURL,
	)
	var labels []string
	for k, v := range msg.CommonLabels {
		if k == "alertname" || k == "severity" {
			continue
		}
		labels = append(labels, k+"="+v)
	}
	text += "labels{" + strings.Join(labels, ",") + "}"

	titleTemplate, err := template.New("title").Parse(`[{{ .Status }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .CommonLabels.alertname }} for {{ .CommonLabels.job }}
      {{- if gt (len .CommonLabels) (len .GroupLabels) -}}
        {{" "}}(
        {{- with .CommonLabels.Remove .GroupLabels.Names }}
          {{- range $index, $label := .SortedPairs -}}
            {{ if $index }}, {{ end }}
            {{- $label.Name }}="{{ $label.Value -}}"
          {{- end }}
        {{- end -}}
        )
      {{- end }}`)
	if err != nil {
		panic(err)
	}

	contentTemplate, err := template.New("content").Parse(`*AlertURL:* <{{ .ExternalURL }}| see more alerts>
{{ range .Alerts -}}
*Alert:* {{if .Annotations.title }}{{ .Annotations.title }} {{ else }}{{ .Labels.alertname}}{{ end }}{{ if .Labels.severity }} - ` + "`{{ .Labels.severity }}`" + `{{ end }}
*PromethusLink:* <{{ .GeneratorURL }}| see promethus source>
*Description:* {{ .Annotations.description }}
*Details:*
{{ range .Labels.SortedPairs }} • *{{ .Name }}:* ` + "`{{ .Value }}`" + `
{{ end }}
{{ end }}`)
	if err != nil {
		panic(err)
	}

	var tpl1 bytes.Buffer
	if err := titleTemplate.Execute(&tpl1, msg); err != nil {
		logrus.Errorf("Unable to parse template: %2", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	title := tpl1.String()

	var tpl2 bytes.Buffer
	if err := contentTemplate.Execute(&tpl2, msg); err != nil {
		logrus.Errorf("Unable to parse template: %2", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	markdown := tpl2.String()
	m := storage.Message{
		Source:    storage.MessageSourceWebhook,
		From:      caller.Name,
		To:        caller.Caps[0],
		Title:     title,
		Color:     getStatusColor(msg.Status),
		Text:      text,
		Markdown:  markdown,
		Timestamp: time.Now().UnixNano(),
	}

	if err := s.store.Save(&m); err != nil {
		logrus.Errorf("Unable to send message: %2", err.Error())
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

func getStatusColor(status string) string {
	switch status {
	case "firing":
		return "#dc3545"
	case "resolved":
		return "#28a745"
	}
	return "#17a2b8"
}
