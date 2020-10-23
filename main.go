package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v32/github"
)

type eventType int

const (
	eventCommit eventType = iota
	eventIssue
	eventPullRequest
)

type event struct {
	createdAt time.Time
	eventType eventType
	message   string
	number    int
}

func flattenString(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

func getCommitEvents(gh *github.Client, owner, repository string) (events []event, err error) {
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	commits, _, err := gh.Repositories.ListCommits(context.Background(), owner, repository, opt)
	if err != nil {
		return nil, err
	}

	for _, commit := range commits {
		events = append(events, event{
			createdAt: *commit.Commit.Author.Date,
			eventType: eventCommit,
			message:   flattenString(*commit.Commit.Message),
			number:    0,
		})
	}
	return events, nil
}

func getIssueEvents(gh *github.Client, owner, repository string) (events []event, err error) {
	opt := &github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	issues, _, err := gh.Issues.ListByRepo(context.Background(), owner, repository, opt)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		var eventType eventType
		if issue.IsPullRequest() {
			eventType = eventPullRequest
		} else {
			eventType = eventIssue
		}
		events = append(events, event{
			createdAt: *issue.CreatedAt,
			eventType: eventType,
			message:   *issue.Title,
			number:    *issue.Number,
		})
	}
	return events, nil
}

func printEvent(e event) {
	var s string
	switch e.eventType {
	case eventCommit:
		s = color.YellowString("    %s %s", e.createdAt.Format("2006-01-02 15:04:05"), e.message)
	case eventIssue:
		s = color.MagentaString("#%2d %s %s", e.number, e.createdAt.Format("2006-01-02 15:04:05"), e.message)
	case eventPullRequest:
		s = color.CyanString("#%2d %s %s", e.number, e.createdAt.Format("2006-01-02 15:04:05"), e.message)
	}
	fmt.Println(s)
}

func run(owner, repository string) error {
	gh := github.NewClient(nil)

	events := make([]event, 0)
	e, err := getCommitEvents(gh, owner, repository)
	if err != nil {
		return err
	}
	events = append(events, e...)

	e, err = getIssueEvents(gh, owner, repository)
	if err != nil {
		return err
	}
	events = append(events, e...)

	sort.Slice(events, func(i, j int) bool {
		return !events[i].createdAt.After(events[j].createdAt)
	})

	for _, e := range events {
		printEvent(e)
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: ghtl [owner] [repository]")
		return
	}

	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
