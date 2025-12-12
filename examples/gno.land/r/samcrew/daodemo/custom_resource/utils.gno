package custom_resource

import (
	"gno.land/p/mason/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/daokit"
)

// initBlogProposals creates initial blog post proposals to demonstrate custom resource functionality
func initBlogProposals() {
	blogPosts := []struct {
		title    string
		content  string
		proposer string
	}{
		{"Welcome to our DAO Blog", "This is our first blog post created through DAO governance!", "g1demo1234567890abcdefghijklmnopqrstuvw1"},
		{"Custom Resources in Action", "This post demonstrates how custom resources work in daokit. Any custom action can be proposed and voted on!", "g1demo1234567890abcdefghijklmnopqrstuvw2"},
		{"Community Updates", "Regular updates from our DAO community will be posted here through democratic proposals.", "g1demo1234567890abcdefghijklmnopqrstuvw3"},
	}

	for _, post := range blogPosts {
		// Calling the custom action
		action := NewPostAction(post.title, post.content)

		req := daokit.ProposalRequest{
			Title:       ufmt.Sprintf("Create blog post: %s", post.title),
			Description: ufmt.Sprintf("This proposal will create a new blog post."),
			Action:      action,
		}
		daoPrivate.Core.Propose(post.proposer, req)
	}
}

// Bypass limitation by adding yourself to the DAO.
// It is necessary to be part of the DAO to create a Proposal.
func AddMember(cur realm) {
	id := daoPrivate.CallerID()
	daoPrivate.Members.AddMember(id, make([]string, 0))
}

// Creates a Proposal to create a new blog post.
// This demonstrates the custom resource functionality of the DAO.
// Parameters:
//   - title: The title of the blog post
//   - content: The content of the blog post
func ProposeNewPost(cur realm, title, content string) {
	// Calling the custom action
	action := NewPostAction(title, content)

	req := daokit.ProposalRequest{
		Title:       ufmt.Sprintf("Create new blog post: %s", title),
		Description: ufmt.Sprintf("This proposal will create a new blog post with title '%s' and content: %s", title, content),
		Action:      action,
	}
	Propose(cur, req)
}

func renderDemo() string {
	s := ""
	s += "# 🔧 Custom Resource Demo\n\n"
	s += "This showcases how to extend your DAO with custom actions.\n\n"
	s += "## ℹ️ How it Works\n\n"
	s += "1. **Join the DAO** using the " + md.Link("AddMember", txlink.Call("AddMember")) + " function.\n\n"
	s += "2. **Create a blog post Proposal** using " + md.Link("ProposeNewPost", txlink.Call("ProposeNewPost")) + " function with these parameters:\n\n"
	s += "- `title`: The title of the blog post\n"
	s += "- `content`: The content of the blog post\n\n"
	s += "3. " + md.Link("Vote", txlink.Call("Vote")) + " on proposals using their ID.\n\n"
	s += "4. " + md.Link("Execute", txlink.Call("Execute")) + " proposals that have passed (met the required voting conditions).\n\n"
	s += "   📝 **Note**: Only proposals that have met the governance conditions can be executed. You need at least **10% of member approval** to execute a proposal.\n\n"
	s += "## 📋 Available Actions\n\n"
	s += "### Member Management\n"
	s += "- " + md.Link("🔗 Add Yourself as Member", txlink.Call("AddMember")) + " - Join the DAO to participate in governance\n"
	s += "### Custom Resource Actions\n"
	s += "- " + md.Link("🔗 Propose New Blog Post", txlink.Call("ProposeNewPost")) + " - Create a proposal to publish a new blog post (demonstrates custom resource)\n\n"
	s += "### Other Demos\n"
	s += "- " + md.Link("🔗 Simple DAO Demo", "/r/samcrew/daodemo/simple_dao") + " - Basic DAO functionality\n"
	s += "- " + md.Link("🔗 Custom Condition Demo", "/r/samcrew/daodemo/custom_condition") + " - Custom governance conditions\n\n"
	s += "*This is a demonstration DAO built with gno.land and daokit*"
	return s
}
