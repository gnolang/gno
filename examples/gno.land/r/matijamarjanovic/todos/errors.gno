package todos

import "errors"

var (
    ErrTokenCannotBeBurned = errors.New("token cannot be burned unless its status is 'done'")
    ErrOnlyCreatorCanMint  = errors.New("only the creator can mint new tasks")
    ErrTokenNotFound       = errors.New("token with specified name not found")
    // Add other errors here if needed
)
