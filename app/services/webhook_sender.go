package services

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

var timeSensitiveDelivery = []time.Duration{
	time.Duration(1) * time.Second,
	time.Duration(3) * time.Second,
	time.Duration(5) * time.Second,
	time.Duration(15) * time.Second,
	time.Duration(30) * time.Second,
	time.Duration(60) * time.Second,
}

func retry(schedule []time.Duration, fn func() error) error {
	var err error
	err = fn()
	if err != nil {
		for _, delay := range schedule {
			err = fn()
			if err == nil {
				return nil
			}
			time.Sleep(delay)
		}
	}
	return err
}

func WebhookSender(destination *url.URL, values *url.Values, schedule []time.Duration) error {
	if destination == nil {
		return fmt.Errorf("URL unconfigured")
	}

	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	err := retry(schedule, func() error {
		res, err := c.PostForm(destination.String(), *values)
		if err == nil && res.StatusCode > 299 {
			return fmt.Errorf("Status Code: %v", res.StatusCode)
		}
		return err
	})

	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			// avoid reporting the URL with potential HTTP auth credentials
			return errors.Wrap(urlErr.Err, "PostForm")
		}
		return errors.Wrap(err, "PostForm")
	}

	return nil
}
