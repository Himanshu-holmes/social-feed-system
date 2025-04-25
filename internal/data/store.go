package data

import (
	"fmt"
	"sort"

	"sync"
	"time"

	"github.com/himanshu-holmes/social-feed-system/internal/models"
)

// InMemoryStore holds all data in memory. Not thread-safe by default for writes
// after initialization, but reads are concurrent-safe.
// For this simulation, we assume data is initialized once.
type InMemoryStore struct {
	Users   map[string]*models.User            // UserID -> User
	Posts   map[string][]*models.Post          // UserID -> Posts by this user
	Follows map[string]map[string]bool         // FollowerID -> {FollowedID: true}
	mu      sync.RWMutex // To protect concurrent access if needed later, though primarily for initialization here
}

func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		Users:   make(map[string]*models.User),
		Posts:   make(map[string][]*models.Post),
		Follows: make(map[string]map[string]bool),
	}
	store.populateMockData()
	return store
}

func (s *InMemoryStore) populateMockData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create Users
	users := []*models.User{
		{ID: "user1", Username: "Alice"},
		{ID: "user2", Username: "Bob"},
		{ID: "user3", Username: "Charlie"},
		{ID: "user4", Username: "David"},
		{ID: "user5", Username: "Eve"},
	}
	for _, u := range users {
		s.Users[u.ID] = u
		s.Posts[u.ID] = []*models.Post{} // Initialize post list for each user
	}

	// Create Posts (Reverse Chronological Order for easier verification)
	now := time.Now()
	postIDCounter := 0
	addPost := func(authorID, content string, timestamp time.Time) {
		postIDCounter++
		post := &models.Post{
			ID:        fmt.Sprintf("post%d", postIDCounter),
			Content:   content,
			Timestamp: timestamp,
			AuthorID:  authorID,
		}
		s.Posts[authorID] = append(s.Posts[authorID], post)
	}

	// User 2 (Bob) Posts
	for i := 0; i < 15; i++ {
		addPost("user2", fmt.Sprintf("Bob's post #%d", 15-i), now.Add(-time.Duration(i*5)*time.Minute))
	}
	// User 3 (Charlie) Posts
	for i := 0; i < 8; i++ {
		addPost("user3", fmt.Sprintf("Charlie's thoughts %d", 8-i), now.Add(-time.Duration(i*12)*time.Minute))
	}
	// User 4 (David) Posts
	for i := 0; i < 5; i++ {
		addPost("user4", fmt.Sprintf("David here, post %d", 5-i), now.Add(-time.Duration(i*30)*time.Minute))
	}
    // User 5 (Eve) Posts - No posts

	// Create Follow relationships
	// User 1 follows User 2, User 3
	s.Follows["user1"] = map[string]bool{
		"user2": true,
		"user3": true,
	}
	// User 2 follows User 1, User 4
	s.Follows["user2"] = map[string]bool{
		"user1": true, // Alice has no posts
		"user4": true,
	}
    // User 3 follows everyone except themself
    s.Follows["user3"] = map[string]bool{
        "user1": true,
        "user2": true,
        "user4": true,
        "user5": true, // Eve has no posts
    }
	// User 4 follows User 2
    s.Follows["user4"] = map[string]bool{
        "user2": true,
    }
    // User 5 follows no one
    s.Follows["user5"] = map[string]bool{}


	// Ensure posts within each user list are sorted (descending) - mock data already is
	sort.Slice(s.Posts["user2"], func(i, j int) bool { return s.Posts["user2"][i].Timestamp.After(s.Posts["user2"][j].Timestamp) })
	sort.Slice(s.Posts["user3"], func(i, j int) bool { return s.Posts["user3"][i].Timestamp.After(s.Posts["user3"][j].Timestamp) })
	sort.Slice(s.Posts["user4"], func(i, j int) bool { return s.Posts["user4"][i].Timestamp.After(s.Posts["user4"][j].Timestamp) })

	fmt.Println("Mock data populated:")
	fmt.Printf("  Users: %d\n", len(s.Users))
	totalPosts := 0
	for _, posts := range s.Posts {
		totalPosts += len(posts)
	}
	fmt.Printf("  Posts: %d\n", totalPosts)
	fmt.Printf("  Follow relationships: %d\n", len(s.Follows))
}

func (s *InMemoryStore) GetUser(userID string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.Users[userID]
	return user, ok
}

func (s *InMemoryStore) GetFollowedUsers(userID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	followedMap, ok := s.Follows[userID]
	if !ok {
		return []string{}
	}
	followedIDs := make([]string, 0, len(followedMap))
	for id := range followedMap {
		followedIDs = append(followedIDs, id)
	}
	return followedIDs
}

func (s *InMemoryStore) GetPostsByUser(userID string) ([]*models.Post,error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to prevent external modification? For this sim, direct slice is ok.
	posts, ok := s.Posts[userID]
	if !ok {
		return []*models.Post{}, fmt.Errorf("user %s not found", userID)
	}
	// Posts are assumed to be sorted descending by time during population
	return posts,nil
}