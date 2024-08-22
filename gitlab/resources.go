package gitlab

import (
	"encoding/json"
	"fmt"
	"os"
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
	CreatedAt time.Time  `json:"created_at"`
}

func (issue Issue) Convert(mappings map[int]string) string {
	author := fmt.Sprintf("%d", issue.AuthorId)
	if ghname, ok := mappings[issue.AuthorId]; ok {
		author = ghname
	}

	return fmt.Sprintf("--- IMPORTED ISSUE ---\n[%s] [@%s]\n\n%s",
		issue.CreatedAt.Format(time.RFC3339),
		author,
		issue.Description)
}

func (issue Issue) Summarize() string {
	return fmt.Sprintf("[%d] [uid=%d] [comments=%d] %s\n", issue.Id, issue.AuthorId, len(issue.Comments), issue.Title)
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

func (c Comment) Convert(mappings map[int]string) string {
	author := fmt.Sprintf("%d", c.AuthorId)
	if ghname, ok := mappings[c.AuthorId]; ok {
		author = ghname
	}
	return fmt.Sprintf("--- IMPORTED COMMENT ---\n[%s] [@%s]\n\n%s",
		c.CreatedAt.Format(time.RFC3339),
		author,
		c.Note)
}

func Parse(path string) ([]*Issue, error) {
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

	return curateIssues(issues), nil
}

func curateIssues(issues []*Issue) []*Issue {
	// sort issues
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Id < issues[j].Id
	})

	// filter comments
	for _, issue := range issues {
		issue.Comments = filterComments(issue.Comments, []string{
			"mentioned in",
			"assigned to",
			"unassigned",
			"changed the description",
			"created branch",
			"changed title",
			"marked the checklist",
			"marked this issue",
		})
	}

	return issues
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
