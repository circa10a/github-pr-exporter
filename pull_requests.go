package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	log "github.com/sirupsen/logrus"
)

type PullRequest struct {
	User           string
	CreatedAt      string
	PullRequestURL string
	Status         string
}

type PullRequests []PullRequest

func (p *PullRequests) filterUserRepos() {
	filteredPullRequests := PullRequests{}
	for _, pr := range *p {
		userNamespace := strings.Split(pr.PullRequestURL, "/")[3]
		if pr.User != userNamespace {
			filteredPullRequests = append(filteredPullRequests, pr)
		}
	}
	*p = filteredPullRequests
}

func (u User) getPullRequests(ctx context.Context, client *github.Client, beginningSearchDate string, filterUserRepos bool) PullRequests {
	userPullRequests := PullRequests{}

	// GitHub Search API only supports strings of 256 chars
	searchString := fmt.Sprintf("type:pr author:%s created:>=%s", u, beginningSearchDate)
	search, _, err := client.Search.Issues(ctx, searchString, &github.SearchOptions{})
	if err != nil {
		log.Error(err)
	}

	for _, result := range search.Issues {
		pr := PullRequest{
			User:           *result.User.Login,
			CreatedAt:      result.CreatedAt.Format(time.RFC3339),
			PullRequestURL: *result.PullRequestLinks.HTMLURL,
			Status:         result.GetState(),
		}
		userPullRequests = append(userPullRequests, pr)
	}

	if filterUserRepos {
		userPullRequests.filterUserRepos()
	}

	return userPullRequests
}
