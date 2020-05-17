package slack

import (
	"net/http"

	"github.com/slack-go/slack"
)

func (c *Client) VerifyWithSignedSecret(header http.Header, getData func() ([]byte, error)) (data []byte, ok bool, err error) {
	var verifier slack.SecretsVerifier
	verifier, err = slack.NewSecretsVerifier(header, c.signedSecret)
	if err != nil {
		return
	}

	data, err = getData()
	if err != nil {
		return
	}

	_, err = verifier.Write(data)
	if err != nil {
		return
	}

	if verifyErr := verifier.Ensure(); verifyErr == nil {
		ok = true
	}

	return
}
