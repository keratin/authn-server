package services

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func WebhookSender(destination *url.URL, values *url.Values) error {
	if destination == nil {
		return fmt.Errorf("URL unconfigured")
	}

	res, err := http.PostForm(destination.String(), *values)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			// avoid reporting the URL with potential HTTP auth credentials
			return errors.Wrap(urlErr.Err, "PostForm")
		}
		return errors.Wrap(err, "PostForm")
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("Status Code: %v", res.StatusCode)
	}

	return nil
}
