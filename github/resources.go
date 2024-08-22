package main

import "github.com/kkentzo/gl-to-gh/gitlab"

type Issue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees"`
	Labels    []string `json:"labels"`
	comments  []*Comment
}

type Comment struct {
	Body string `json:"body"`
}

func New(glIssue *gitlab.Issue, mappings map[int]string, labels []string) *Issue {
	issue := &Issue{
		Title:    glIssue.Title,
		Body:     glIssue.Convert(mappings),
		Labels:   labels,
		comments: []*Comment{},
	}
	for _, glComment := range glIssue.Comments {
		comment := &Comment{Body: glComment.Convert(mappings)}
		issue.comments = append(issue.comments, comment)
	}
	return issue
}
