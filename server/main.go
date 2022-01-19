package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/manpreet1130/blog/protos/blogpb"
	"github.com/manpreet1130/blog/server/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

type Post struct {
	gorm.Model
	Post_ID uuid.UUID
	Title   string
	Content string
	Author  string
}

var db *gorm.DB

func init() {
	database.Connect()
	db = database.GetDB()
	db.AutoMigrate(&Post{})
}

func (s *server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()
	fmt.Println("Received:", blog)

	post := Post{}

	db.Where("Title = ?", blog.GetTitle()).Find(&post)

	if post.Title != "" {
		status.Error(codes.AlreadyExists, "blogpost with following title already exists")
		return nil, errors.New("Post with the following title already exists.")
	}

	newPost := &Post{
		Post_ID: uuid.New(),
		Title:   blog.GetTitle(),
		Content: blog.GetContent(),
		Author:  blog.GetAuthor(),
	}

	db.Create(newPost)

	resp := &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:      uint32(newPost.ID),
			Title:   newPost.Title,
			Content: newPost.Content,
			Author:  newPost.Author,
		},
	}

	log.Printf("Added new blog post: %v", newPost)
	return resp, nil
}

func (s *server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	title := req.GetTitle()
	author := req.GetAuthor()

	blogPost := Post{}
	db.Where(&Post{Title: title, Author: author}).Find(&blogPost)

	if blogPost.Title == "" && blogPost.Author == "" {
		status.Error(codes.InvalidArgument, "invalid title provided, no such blog post")
		return nil, errors.New("invalid title and/or author input")
	}

	res := &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:      uint32(blogPost.ID),
			Title:   blogPost.Title,
			Content: blogPost.Content,
			Author:  blogPost.Author,
		},
	}
	log.Printf("Sent Response: %v\n", res)
	return res, nil
}

func (s *server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	title := req.GetTitle()
	author := req.GetAuthor()
	content := req.GetContent()

	post := Post{}

	db.Where(&Post{Title: title, Author: author}).Find(&post)

	if post.Title == "" && post.Author == "" {
		status.Error(codes.InvalidArgument, "could not find post with given title and author")
		return nil, errors.New("invalid title and/or author input")
	}

	post.Content = content
	db.Save(&post)

	log.Printf("Updated blog with title %v and author %v", title, author)
	resp := &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Title:   post.Title,
			Author:  post.Author,
			Content: post.Content,
		},
	}

	return resp, nil
}

func (s *server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	title := req.GetTitle()
	author := req.GetAuthor()

	post := Post{}

	db.Where(&Post{Title: title, Author: author}).Delete(&post)

	log.Printf("Deleted blog with title %v and author %v", title, author)
	resp := &blogpb.DeleteBlogResponse{
		Confirmation: fmt.Sprintf("Blog with title %v and author %v was successfully deleted.", title, author),
	}

	return resp, nil

}

func (s *server) ListBlogs(req *blogpb.ListBlogsRequest, stream blogpb.BlogService_ListBlogsServer) error {
	posts := []Post{}
	db.Find(&posts)

	for _, post := range posts {
		log.Println("Sending: ", post)
		resp := &blogpb.ListBlogsResponse{
			Blog: &blogpb.Blog{
				Title:   post.Title,
				Author:  post.Author,
				Content: post.Content,
			},
		}
		if err := stream.Send(resp); err != nil {
			return status.Error(codes.Internal, "could not setup server stream")
		}
	}

	log.Println("That's all the blogs present")
	return nil

}

func main() {
	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("cannot listen on port 8080")
	}

	blogpb.RegisterBlogServiceServer(grpcServer, &server{})

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt)

	go func() {
		fmt.Println("Server started...")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("could not create server.")
		}
	}()

	<-sigChan
	fmt.Println("Received interrupt signal, commencing graceful shutdown.")
	grpcServer.Stop()
	listener.Close()
	fmt.Println("Shutdown complete, goodbye.")
}
