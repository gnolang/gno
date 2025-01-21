package mux

import (
    "fmt"
    "testing"
)

func TestRequest_GetVar(t *testing.T) {
    cases := []struct {
        handlerPath    string
        reqPath        string
        getVarKey      string
        expectedOutput string
    }{
        
        {"users/{id}", "users/123", "id", "123"},
        {"users/123", "users/123", "id", ""}, 
        {"users/{id}", "users/123", "nonexistent", ""}, 

        {"users/{userId}/posts/{postId}", "users/123/posts/456", "userId", "123"},
        {"users/{userId}/posts/{postId}", "users/123/posts/456", "postId", "456"},
        
        // Wildcards
		{"*", "users/123", "*", "users/123"},
		{"*", "users/123/posts/456", "*", "users/123/posts/456"},
		{"*", "users/123/posts/456/comments/789", "*", "users/123/posts/456/comments/789"},
		
		{"users/*", "users/john/posts", "*", "john/posts"},
		{"users/*/comments", "users/jane/comments", "*", "jane/comments"},
		{"api/*/posts/*", "api/v1/posts/123", "*", "v1/posts/123"},
        
        // Mixed wildcards and parameters
        {"api/{version}/*", "api/v1/user/settings", "version", "v1"},
        
        }

    for _, tt := range cases {
        name := fmt.Sprintf("%s-%s-%s", tt.handlerPath, tt.reqPath, tt.getVarKey)
        t.Run(name, func(t *testing.T) {
            req := &Request{
                HandlerPath: tt.handlerPath,
                Path:       tt.reqPath,
            }

            output := req.GetVar(tt.getVarKey)
            if output != tt.expectedOutput {
                t.Errorf("GetVar(%q) = %q, want %q (handler: %q, path: %q)", 
                    tt.getVarKey, output, tt.expectedOutput, tt.handlerPath, tt.reqPath)
            }
        })
    }
}