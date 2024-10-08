package gitlab

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Issue struct {
	Id          int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorId    int    `json:"author_id"`
	Assignees   []struct {
		UserId int `json:"user_id"`
	} `json:"issue_assignees"`
	Comments  []*Comment `json:"notes"`
	State     string     `json:"state"`
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  time.Time  `json:"closed_at"`
}

func (issue Issue) Convert(mappings map[int]string, replPatterns map[string]string) (string, error) {
	author := fmt.Sprintf("%d", issue.AuthorId)
	if ghname, ok := mappings[issue.AuthorId]; ok {
		author = "@" + ghname
	}

	description, err := ReplaceAll(issue.Description, replPatterns)
	if err != nil {
		return "", err
	}

	closedAt := "<n/a>"
	if issue.IsClosed() {
		closedAt = issue.ClosedAt.Format(time.RFC3339)
	}

	return fmt.Sprintf("\nISSUE IMPORTED FROM GITLAB [id=%d] [state=%s]\ncreated: `%s`\nclosed: `%s`\noriginal author: %s\ncomments: %d\n\n---\n\n%s",
		issue.Id,
		issue.State,
		issue.CreatedAt.Format(time.RFC3339),
		closedAt,
		author,
		len(issue.Comments),
		description), nil
}

func (issue Issue) Summarize() string {
	return fmt.Sprintf("[%d] [uid=%d] [comments=%d] %s\n", issue.Id, issue.AuthorId, len(issue.Comments), issue.Title)
}

func (issue Issue) IsClosed() bool {
	return !issue.ClosedAt.IsZero()
}

type Comment struct {
	Note         string    `json:"note"`
	AuthorId     int       `json:"author_id"`
	DiscussionId string    `json:"discussion_id"`
	CreatedAt    time.Time `json:"created_at"`
	Author       struct {
		Name string `json:"name"`
	} `json:"author"`
}

func (c Comment) Convert(replPatterns map[string]string) (string, error) {
	description, err := ReplaceAll(c.Note, replPatterns)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\nCOMMENT IMPORTED FROM GITLAB\ncreated: `%s`\noriginal author: %s\n\n---\n\n%s",
		c.CreatedAt.Format(time.RFC3339),
		c.Author.Name,
		description), nil
}

// Parse the ndjson file specified by path
// and return the issues ordered by ID (asc)
// Also apply any filters to all comments
func Parse(path string, commentExclusionFilter []string) ([]*Issue, error) {
	issues := []*Issue{}

	src, err := os.Open(path)
	if err != nil {
		return issues, err
	}
	defer src.Close()

	decoder := json.NewDecoder(src)

	for decoder.More() {
		issue := &Issue{}
		if err := decoder.Decode(issue); err != nil {
			return issues, err
		}
		issues = append(issues, issue)
	}

	return curateIssues(issues, commentExclusionFilter), nil
}

func curateIssues(issues []*Issue, commentExclusionFilter []string) []*Issue {
	// sort issues
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Id < issues[j].Id
	})

	// filter and sort comments
	for _, issue := range issues {
		issue.Comments = filterComments(issue.Comments, commentExclusionFilter)
		sort.Slice(issue.Comments, func(i, j int) bool {
			return issue.Comments[i].CreatedAt.Before(issue.Comments[j].CreatedAt)
		})
	}

	return issues
}

func ReplaceAll(src string, patterns map[string]string) (string, error) {
	var err error
	for expr, repl := range patterns {
		src, err = Replace(src, expr, repl)
		if err != nil {
			return src, fmt.Errorf("failed to apply exp=%s: %v", expr, err)
		}
	}
	return src, nil
}

// return a copy of src, replacing matches of the regex expr with the replacement string repl.
// Inside repl, $ signs are interpreted as the texts of the submatches
// See test file for example(s)
func Replace(src, expr, repl string) (string, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return src, fmt.Errorf("error compiling the regex expression %s: %v", expr, err)
	}
	return re.ReplaceAllString(src, repl), nil
}

func Users(issues []*Issue) []int {
	users := map[int]bool{}
	for _, issue := range issues {
		for _, assignee := range issue.Assignees {
			users[assignee.UserId] = true
		}
	}

	ids := []int{}
	for id := range users {
		ids = append(ids, id)
	}
	return ids
}

func filterComments(comments []*Comment, blacklist []string) []*Comment {
	filtered := []*Comment{}
	for _, comment := range comments {
		rejected := false
		for _, prefix := range blacklist {
			if strings.HasPrefix(comment.Note, prefix) {
				rejected = true
				break
			}
		}
		if rejected {
			continue
		}
		filtered = append(filtered, comment)
	}
	return filtered
}

func printStats(issues []*Issue) {
	fmt.Printf("Issues: %d\n", len(issues))
	nc := 0
	for _, issue := range issues {
		nc += len(issue.Comments)
	}
	fmt.Printf("Comments: %d\n", nc)

	fmt.Print("User IDs: ")
	for _, uid := range Users(issues) {
		fmt.Printf("%d ", uid)
	}
	fmt.Println()
}
