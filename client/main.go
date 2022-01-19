package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/manpreet1130/blog/protos/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("could not connect to server")
	}

	client := blogpb.NewBlogServiceClient(conn)

	// addBlog(client, "Another title", "Some author", "Some content.")

	// readBlog(client, "Life", "The One and Only")

	// updateBlog(client, "Some title", "Some author", "Updated some content")

	// deleteBlog(client, "Some title", "Some author")

	// getBlogs(client)
}

func addBlog(client blogpb.BlogServiceClient, title string, author string, content string) {
	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			Title:   title,
			Content: content,
			Author:  author,
		},
	}

	resp, err := client.CreateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("could not get response back, %v", err)
	}

	fmt.Println("Response:", resp)
}

func readBlog(client blogpb.BlogServiceClient, title string, author string) {
	req := blogpb.ReadBlogRequest{
		Title:  title,
		Author: author,
	}

	resp, err := client.ReadBlog(context.Background(), &req)
	if err != nil {
		log.Fatalf("could not get blog with title %v and author %v\n", title, author)
	}

	fmt.Println("Title:", resp.GetBlog().GetTitle())
	fmt.Println("by:", resp.GetBlog().GetAuthor())
	fmt.Println("Content:", resp.GetBlog().GetContent())
}

func updateBlog(client blogpb.BlogServiceClient, title string, author string, content string) {
	req := blogpb.UpdateBlogRequest{
		Title:   title,
		Author:  author,
		Content: content,
	}

	resp, err := client.UpdateBlog(context.Background(), &req)
	if err != nil {
		log.Fatalf("could not update blog with title %v and author %v\n", title, author)
	}

	fmt.Println("Title:", resp.GetBlog().GetTitle())
	fmt.Println("by:", resp.GetBlog().GetAuthor())
	fmt.Println("Content:", resp.GetBlog().GetContent())
}

func deleteBlog(client blogpb.BlogServiceClient, title string, author string) {
	req := blogpb.DeleteBlogRequest{
		Title:  title,
		Author: author,
	}

	resp, err := client.DeleteBlog(context.Background(), &req)
	if err != nil {
		log.Fatalf("could not delete blog with title %v and author %v\n", title, author)
	}

	fmt.Println(resp.GetConfirmation())
}

func getBlogs(client blogpb.BlogServiceClient) {
	req := &blogpb.ListBlogsRequest{}
	responseStream, err := client.ListBlogs(context.Background(), req)
	if err != nil {
		log.Fatal("could not receive the response stream")
	}

	for {
		message, err := responseStream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Stream closed, exitting")
				break
			}
			log.Fatal("error while receiving message from stream")
		}

		fmt.Println("Title: ", message.GetBlog().GetTitle())
		fmt.Println("Author: ", message.GetBlog().GetAuthor())
		fmt.Println("Content: ", message.GetBlog().GetContent())
	}

	fmt.Println("Received all present blogs")
}
