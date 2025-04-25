package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"sync"

	"github.com/himanshu-holmes/social-feed-system/internal/data"
	"github.com/himanshu-holmes/social-feed-system/internal/models"
	postV1 "github.com/himanshu-holmes/social-feed-system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultPort = "6001"
	timelinelimit = 20
)

type Server struct {
	store *data.InMemoryStore
	postV1.UnimplementedPostServiceServer
}

func NewServer() *Server {
	store := data.NewInMemoryStore()
	return &Server{
		store: store,
	}
}

func (s *Server) Run() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = defaultPort
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen : %v", err)
	}
	grpcServer := grpc.NewServer()
	postV1.RegisterPostServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	log.Printf("Starting PostService server on port %s", port)
	go func() {
		grpcServer.Serve(listener)
	}()
}

func (s *Server) ListPostsByUser(ctx context.Context, req *postV1.ListPostsRequest) (*postV1.ListPostsResponse, error) {
	userId := req.UserId
	// 1. Get the list of users followed by the given userId
	followedUserIds := s.store.GetFollowedUsers(userId)
	if len(followedUserIds) == 0 {
		log.Printf("User %s folows no one ", userId)
		return &postV1.ListPostsResponse{}, nil
	}

	//2. Fetch posts for each followed user concurrently
	var wg sync.WaitGroup
	postChan := make(chan []*models.Post, len(followedUserIds))
	errChan := make(chan error, len(followedUserIds))
	for _, followedId := range followedUserIds {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			posts, err := s.store.GetPostsByUser(id)
			if err != nil {
				log.Printf("Error fetching posts for user %s, %v", id, err)
				select {
				case errChan <- fmt.Errorf("error fetching posts for user %s: %v", id, err):

				case <-ctx.Done():
					log.Printf("Context cancelled while fetching posts for user %s", id)

				}
				return
			}
			if len(posts) > 0 {
				select {
				case postChan <- posts:

				case <-ctx.Done():
					log.Printf("Context cancelled while sending posts for user %s", id)
				}
			}

		}(followedId)
	}
	go func() {
		wg.Wait()
		close(postChan)
		close(errChan)
	}()

	// 3. Aggregate the results and also handle errors
	allPosts := []*models.Post{}
	var encounteredError error

	// loop untils both channels are closed
	for postChan != nil || errChan != nil {
		select {
		case posts, ok := <-postChan:
			if !ok {
				postChan = nil
				continue
			}
			allPosts = append(allPosts, posts...)

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			log.Printf("Error fetching posts: %v", err)
			encounteredError = err

		case <-ctx.Done():
			log.Printf("Context cancelled while aggregating posts")
			go func(){
				for range postChan {
					// drain the channel

				}
				for range errChan {
					// drain the channel
				}
			}()
			return nil, fmt.Errorf("timeline aggregation cancelled: %w", ctx.Err())
		}
		

	}
	log.Printf("Aggregated %d posts from %d followed users",len(allPosts),len(followedUserIds))
	// 4. sort aggregated posts by timestamp (descending)
	sort.SliceStable(allPosts,func(i, j int) bool {
		return allPosts[i].Timestamp.After(allPosts[j].Timestamp)
	})
	// 5. Limit to the most recent N posts
	if len(allPosts) > timelinelimit {
		allPosts = allPosts[:timelinelimit]
	}

	log.Printf("Returning top %d posts for timeline.",len(allPosts))
	// 6. Convert to protobuf response
	response := &postV1.ListPostsResponse{
		Posts: make([]*postV1.Post, len(allPosts)),
	}
	for i, post := range allPosts {
		response.Posts[i] = &postV1.Post{
			Id:        post.ID,
			Content:   post.Content,
			Timestamp: timestamppb.New(post.Timestamp),
			AuthorId:  post.AuthorID,
		}
	}
	return response, encounteredError
}
