package okta

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestRateLimitThrottleCheckIsEndpoint(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	apiV1AppsEndpoints := newRateLimitThrottle([]string{
		// the following endpoints share the same rate limit
		`/api/v1/apps$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/groups$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/groups/(?P<GroupID>[[:alnum:]]+)$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/users$`,
	}, 5)
	tests := map[string]bool{
		"/api/v1/apps":                true,
		"/api/v1/apps/123/groups":     true,
		"/api/v1/apps/123/users":      true,
		"/api/v1/apps/123/groups/456": true,
		"/api/v1/apps/123":            false,
	}
	for path, expected := range tests {
		t.Run(path, func(t *testing.T) {
			if apiV1AppsEndpoints.checkIsEndpoint(path) != expected {
				t.Errorf("Path %v got incorrectly interpreted.", path)
			}
		})
	}
}

func TestRateLimitThrottlePreRequestHook(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	apiV1AppsEndpoints := newRateLimitThrottle([]string{
		// the following endpoints share the same rate limit
		`/api/v1/apps$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/groups$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/groups/(?P<GroupID>[[:alnum:]]+)$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/users$`,
	}, 5)
	apiV1AppsEndpoints.maxRequests = 40
	apiV1AppsEndpoints.rateLimit = 25
	apiV1AppsEndpoints.rateLimitResetTime = time.Now().Add(time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	allowedRequests := apiV1AppsEndpoints.rateLimit * apiV1AppsEndpoints.maxRequests / 100.
	for requestNo := 1; requestNo <= allowedRequests; requestNo++ {
		err := apiV1AppsEndpoints.preRequestHook(ctx, "/api/v1/apps")
		if apiV1AppsEndpoints.noOfRequestsMade != requestNo {
			t.Errorf(
				"Incorrect request count after request number %v. Expected %v, got %v.",
				requestNo,
				requestNo,
				apiV1AppsEndpoints.noOfRequestsMade,
			)
		}
		if err != nil {
			t.Errorf("Unexpected error after request number %v: %v", requestNo, err)
		}
	}
	if err := apiV1AppsEndpoints.preRequestHook(ctx, "/api/v1/apps"); err != context.Canceled {
		t.Errorf("Expected %v error, got %v.", context.Canceled, err)
	}
}

func TestRateLimitThrottlePostRequestHook(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	apiV1AppsEndpoints := newRateLimitThrottle([]string{
		// the following endpoints share the same rate limit
		`/api/v1/apps$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/groups$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/groups/(?P<GroupID>[[:alnum:]]+)$`,
		`/api/v1/apps/(?P<AppID>[[:alnum:]]+)/users$`,
	}, 5)

	request := http.Request{
		URL: &url.URL{
			Path: "/api/v1/apps",
		},
	}
	rateLimit := 25
	rateLimitResetTime := time.Now()
	responseHeaders := http.Header{}
	responseHeaders.Add("X-Rate-Limit-Limit", fmt.Sprintf("%v", rateLimit))
	responseHeaders.Add("X-Rate-Limit-Reset", fmt.Sprintf("%v", rateLimitResetTime.Unix()))
	response := http.Response{
		Request: &request,
		Header:  responseHeaders,
	}
	apiV1AppsEndpoints.postRequestHook(&response)
	if apiV1AppsEndpoints.rateLimit != rateLimit {
		t.Errorf(
			"Rate limit header got parsed incorrectly. Expected %v, got: %v.",
			rateLimit,
			apiV1AppsEndpoints.rateLimit,
		)
	}
	if apiV1AppsEndpoints.rateLimitResetTime.Equal(rateLimitResetTime) {
		t.Errorf(
			"Rate limit reset time header got parsed incorrectly. Expected %v, got: %v.",
			rateLimitResetTime,
			apiV1AppsEndpoints.rateLimitResetTime,
		)
	}
}
