package server

import (
	"context"
	"testing"
	

	"github.com/himanshu-holmes/social-feed-system/internal/data"
	"github.com/himanshu-holmes/social-feed-system/proto"
	
)

func TestListPostsByUser(t *testing.T) {
	server := &Server{store: data.NewInMemoryStore()}

	tests := []struct {
		userID         string
		expectedCount  int
		description    string
	}{
		{"user1", 20, "user1 follows user2 (15 posts) + user3 (8 posts) = 23 -> capped to 20"},
		{"user2", 5, "user2 follows user1 (0) + user4 (5) = 5"},
		{"user3", 20, "user3 follows user1,2,4,5 -> total 28 -> capped to 20"},
		{"user4", 15, "user4 follows user2 -> 15 posts"},
		{"user5", 0, "user5 follows no one"},
	}

	for _, tt := range tests {
		resp, err := server.ListPostsByUser(context.Background(), &proto.ListPostsRequest{
			UserId: tt.userID,
		})
		if err != nil {
			t.Errorf("unexpected error for %s: %v", tt.description, err)
		}
		if len(resp.Posts) != tt.expectedCount {
			t.Errorf("failed %s: expected %d posts, got %d", tt.description, tt.expectedCount, len(resp.Posts))
		}

		// Check reverse chronological order
		for i := 1; i < len(resp.Posts); i++ {
			curr := resp.Posts[i].Timestamp.AsTime()
			prev := resp.Posts[i-1].Timestamp.AsTime()
			if prev.Before(curr) {
				t.Errorf("posts are not in descending order by time for %s", tt.description)
			}
		}
	}
}

func TestContextCancellation(t *testing.T) {
	server := &Server{store: data.NewInMemoryStore()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	_, err := server.ListPostsByUser(ctx, &proto.ListPostsRequest{
		UserId: "user1",
	})

	if err == nil {
		t.Errorf("expected context cancellation error but got nil")
	}
}
