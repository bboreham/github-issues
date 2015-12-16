package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/google/go-github/github"
)

var iTemplate = template.Must(template.New("issues").Parse(
	"{{if .PullRequestLinks}}{{if .PRMerged}}- PR {{.Title}} #{{.Number}}\n" +
		"{{end}}" +
		"{{else}}" +
		"- {{.Title}} #{{.Number}} {{.State}}\n" +
		"{{end}}",
))

func main() {
	client := github.NewClient(nil)

	owner, repo := "weaveworks", "weave"
	milestone := "1.3.2"
	milestones, _, err := client.Issues.ListMilestones(owner, repo, &github.MilestoneListOptions{State: "all"})
	if err != nil {
		log.Fatal(err)
	}
	milestoneNumber := ""
	for _, m := range milestones {
		if m.Title != nil && *m.Title == milestone && m.Number != nil {
			milestoneNumber = fmt.Sprintf("%d", *m.Number)
			break
		}
	}
	if milestoneNumber == "" {
		log.Fatal("Unable to find milestone", milestone)
	}

	issues, _, err := client.Issues.ListByRepo(owner, repo, &github.IssueListByRepoOptions{Milestone: milestoneNumber, State: "all", ListOptions: github.ListOptions{PerPage: 999}})
	if err != nil {
		log.Fatal(err)
	}

	for _, issue := range issues {
		wrapper := struct {
			github.Issue
			PR       *github.PullRequest
			PRMerged bool
		}{Issue: issue}
		if issue.PullRequestLinks != nil {
			wrapper.PR, _, err = client.PullRequests.Get(owner, repo, *issue.Number)
			if err != nil {
				log.Fatal(err)
			}
			wrapper.PRMerged = *wrapper.PR.Merged
		}

		iTemplate.Execute(os.Stdout, wrapper)
	}
}
