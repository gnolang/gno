# Social Follow System

## Overview
The Social Network System is designed to facilitate user interactions within a social network. It allows users to follow and unfollow each other, as well as retrieve information about followers and followed users.

## Modules

### 1. User Management
- This module manages user information within the social follow.
- Each user is represented by a `User` struct containing their address and lists of followers and followed users.

### 2. Following Functionality
- Users can follow and unfollow other users.
- When a user follows another user, they are added to the followed user's list of followers, and the followed user is added to the user's list of followed users.
- When a user unfollows another user, they are removed from the followed user's list of followers, and the followed user is removed from the user's list of followed users.

## Data Structures

### User
```go
type User struct {
	address   std.Address
	followers *avl.Tree // std.Address -> *User
	followeds *avl.Tree // std.Address -> *User
}
```

## Functions

### User Management
- `Followers(page, pageSize int) []std.Address `: Returns a list of addresses of users following the user.
- `Followed(page, pageSize int) []std.Address `: Returns a list of addresses of users whom the user is following.
- `FollowedCount(addr std.Address) uint`: Returns the number of users being followed by the user with the given address.
- `FollowersCount(addr std.Address) uint`: Returns the number of users following the user with the given address.

### Following Functionality
- `Follow(user *User)`: Adds the given user to the list of users being followed by the user.
- `Unfollow(user *User)`: Removes the given user from the list of users being followed by the user.

## Realm Configuration Process
- Users are managed within the social network system using the provided functionality.
- The system utilizes an AVL tree data structure to efficiently store and retrieve user information.

## Usage
- Users interact with the system by following or unfollowing other users.
- User information is maintained and updated dynamically as users follow and unfollow each other.
