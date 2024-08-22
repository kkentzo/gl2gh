package github

import (
	"encoding/json"
	"fmt"

	"github.com/kkentzo/gl-to-gh/gitlab"
)

type Issue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees"`
	Labels    []string `json:"labels"`
	comments  []*Comment
}

func New(glIssue *gitlab.Issue, mappings map[int]string, labels []string, replPatterns map[string]string) (*Issue, error) {
	body, err := glIssue.Convert(mappings, replPatterns)
	if err != nil {
		return nil, err
	}
	issue := &Issue{
		Title:     glIssue.Title,
		Body:      body,
		Labels:    labels,
		Assignees: FindAssignees(glIssue, mappings),
		comments:  []*Comment{},
	}
	for _, glComment := range glIssue.Comments {
		body, err = glComment.Convert(replPatterns)
		if err != nil {
			return nil, err
		}
		comment := &Comment{Body: body}
		issue.comments = append(issue.comments, comment)
	}
	return issue, nil
}

func FindAssignees(glIssue *gitlab.Issue, mappings map[int]string) []string {
	assignees := []string{}
	for _, assignee := range glIssue.Assignees {
		if user, ok := mappings[assignee.UserId]; ok {
			assignees = append(assignees, user)
		}
	}
	return assignees
}

func (issue *Issue) Path(repo string) string {
	return fmt.Sprintf("/repos/%s/issues", repo)
}

func (issue *Issue) Post(client *Client, repo string) error {
	// serialize the issue
	body, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("failed to serialize issue: %v\nThe problematic issue is:\n%v\n", err, issue)
	}
	res, err := client.Post(urljoin(apiEndpoint, issue.Path(repo)), body)
	if err != nil {
		return fmt.Errorf("error posting issue: %v", err)
	}
	// figure out the URL for posting the comments
	response := struct {
		CommentsURL string `json:"comments_url"`
	}{}
	if err := json.Unmarshal(res, &response); err != nil {
		return fmt.Errorf("error parsing issue response body: %v", err)
	}

	// OK, let's post the comments now
	for _, comment := range issue.comments {
		// serialize the comment
		body, err = json.Marshal(comment)
		if err != nil {
			return fmt.Errorf("error serializing comment: %v\nThe problematic comment is:\n%v\n", err, comment)
		}
		// post the comment
		_, err = client.Post(response.CommentsURL, body)
		if err != nil {
			return fmt.Errorf("error posting comment: %v", err)
		}
	}

	return err
}

type Comment struct {
	Body string `json:"body"`
}
