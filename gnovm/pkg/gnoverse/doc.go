// Package gnoverse provides a sandbox environment, referred to as "gnoverse,"
// for running the 'gnovm' with support for managing multiple users and their
// states. The gnoverse allows users to interact with a dynamic and living
// environment where they can perform actions within a controlled and isolated
// space.
//
// This package is designed to be versatile and can be used in various scenarios:
//
//  1. Mocking Gno machine with persistency, multi-user support and an in-memory
//     database for testing purposes.
//
//  2. Creating blockchain-less multi-user Gno servers.
//     Gnoverse can serve as a foundation to create multi-user Gno servers
//     without the need for a blockchain or specific network protocols like HTTP
//     APIs, SSH servers, or full CLI experiences. Each user within the gnoverse
//     can have their state and perform actions, offering a realistic server-like
//     experience for users.
//
// Usage:
// The main entry point to the package is the Sandbox type, which represents the
// gnoverse instance.
//
//	// Create a new gnoverse instance.
//	sandbox := gnoverse.NewSandbox()
//
//	// Manage user interactions, state, and actions within the gnoverse.
//	user1 := sandbox.CreateUser("Alice")
//	user2 := sandbox.CreateUser("Bob")
//
//	// Perform actions as users in the gnoverse.
//	user1.DoAction("jump")
//	user2.DoAction("run")
//
//	// Retrieve user state and information.
//	state := user1.GetState()
//	name := user2.GetName()
//
//	// Extend and customize the gnoverse as needed.
//
// Note:
// The 'gnoverse' package is designed as a foundation for managing the 'gnovm'
// environment with support for multiple users and their states. It does not
// include any specific interfaces for exposing a port or providing HTTP, SSH,
// or CLI server functionalities. Instead, it focuses on providing a sandbox
// environment for interacting with the 'gnovm.'
package gnoverse
