package github

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	if glIssue.IsClosed() {
		labels = append(labels, "closed")
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

func NewPlaceholder(labels []string) *Issue {
	return &Issue{
		Title:     "[DELETED GITLAB ISSUE]",
		Body:      "This issue was created during the import of gitlab issues in order to preserve the ID ordering of gitlab issue IDs. In reality, it represents a deleted gitlab issue.",
		Assignees: []string{},
		Labels:    labels,
		comments:  []*Comment{},
	}
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

func (issue *Issue) Comments() []*Comment {
	return issue.comments
}

func (issue *Issue) Post(client *Client, repo string) error {
	// serialize the issue
	body, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("failed to serialize issue: %v\nThe problematic issue is:\n%v\n", err, issue)
	}
	// prepare the request
	req, err := client.NewRequest(http.MethodPost, urljoin(apiEndpoint, issue.Path(repo)), body)
	if err != nil {
		return fmt.Errorf("error preparing the request: %v", err)
	}
	resBody, err := client.Do(req, http.StatusCreated)
	if err != nil {
		return fmt.Errorf("request failed: %v\nResponse Body=%s", err, string(resBody))
	}

	// figure out the URL for posting the comments
	response := struct {
		CommentsURL string `json:"comments_url"`
	}{}
	if err := json.Unmarshal(resBody, &response); err != nil {
		return fmt.Errorf("error parsing issue response body: %v", err)
	}

	return err
}

type Comment struct {
	Body string `json:"body"`
}

func (comment *Comment) Path(repo string, issueId int) string {
	return fmt.Sprintf("/repos/%s/issues/%d/comments", repo, issueId)
}

func (comment *Comment) Post(client *Client, repo string, issueId int) error {
	// serialize the comment
	body, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("error serializing comment: %v\nThe problematic comment is:\n%v\n", err, comment)
	}
	// post the comment
	req, err := client.NewRequest(http.MethodPost, urljoin(apiEndpoint, comment.Path(repo, issueId)), body)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %v", err)
	}
	if _, err = client.Do(req, http.StatusCreated); err != nil {
		return fmt.Errorf("error posting comment: %v", err)
	}

	return nil
}
