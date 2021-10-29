---
description: Add comment on Blog
order: 2
---

# Add comment to Blog post Blockchain 

In this tutorial, you will create a new message module called comment. The module will let you read and write comments to an existing blog blockchain.

You can only add comments to post which are no older than 100 Blocks. 

**Note:** This value has been hard coded to a low number for rapid testing. You can increase it to a greater number to achieve longer period of time before commenting is not allowed.

### Prerequisites:

- This tutorial is an extension of previously written blog tutorial. Make sure you complete that first before proceeding with this tutorial.
- This tutorial also assumes basic knowledge of blog tutorial implementation.
- Make sure you are inside the `blog` directory created in the previous blog tutorial.

## Create a new message called comment

To create a new message module for comment, use the `message` command:

```bash
starport scaffold message create-comment blogID:int title body
```

The `message` commands accepts `blogID` and a list of fields (`title` and `body` as arguments )
Here, `blogID` is the reference to previously created blog posts.

The `message` command has created and modified several files:

----------->>> TODO

As always, start with a proto file. Inside the `proto/blog/tx.proto` file, the `MsgCreateComment` message has been created. Edit the file to define the id for `message MsgCreateCommentResponse`:

```go
message MsgCreateComment {
  string creator = 1;
  uint64 id = 2;
  string title = 3;
  string body = 4;
  uint64 blogID = 5;
  int64 createdAt = 6;
}

message MsgCreateCommentResponse {
  uint64 id = 1;
}
```

First, define a Cosmos SDK message type with proto `message`. The `MsgCreateComment` has three fields: creator, title, body, blogID and createdAt. Since the purpose of the `MsgCreateComment` message is to create new comments in the store, the only thing the message needs to return is an ID of a created comments. The `CreateComment` rpc was already added to the `Msg` service:

```go
  rpc CreateComment(MsgCreateComment) returns (MsgCreateCommentResponse);
```

Next, look at the `x/blog/handler.go` file. Starport has added a `case` to the `switch` statement inside the `NewHandler` function. This switch statement is responsible for routing messages and calling specific keeper methods based on the type of the message. `case *types.MsgCreateComment` has been added along with `case *types.MsgCreatePost`

```go
func NewHandler(k keeper.Keeper) sdk.Handler {
	//...
		switch msg := msg.(type) {
		//...
		case *types.MsgCreateComment:
			res, err := msgServer.CreateComment(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
			//...
		}
	}
}
```

The `case *types.MsgCreateComment` statement handles messages of type `MsgCreateComment`, calls the `CreateComment` method, and returns back the response.

Every module has a handler function like this to process messages and call keeper methods.

## Process Messages

In the newly scaffolded `x/blog/keeper/msg_server_create_comment.go` file, you can see a placeholder implementation of the `CreateComment` function. Right now it does nothing and returns an empty response. For your blog chain, you want the contents of the message (title and body) to be written to the state as a new comment.

You need to do two things:

- Create a variable of type `Comment` with title and body from the message
- Check if the the comment posted for the respective blog id exists and comment is not older than 100 blocks.
- Append this `Comment` to the store

```go
func (k msgServer) CreateComment(goCtx context.Context, msg *types.MsgCreateComment) (*types.MsgCreateCommentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	BlockHeight := CreatedAt() // Invoke method to get the block height of Post
	PostID := CreatedId() // Invoke method to get the Post ID

	// Create variable of type comment
	var comment = types.Comment{
		Creator: msg.Creator,
		Id:      msg.Id,
		Body:    msg.Body,
		Title:   msg.Title,
		BlogID:  msg.BlogID,
		CreatedAt: ctx.BlockHeight(),
	}

	// Check if the Post Exists for which a comment is being created
	if comment.BlogID > PostID {
		return nil, sdkerrors.Wrapf(types.ErrID, "Post Blog Id %d does not exist for which comment with Blog Id %d was made", PostID, comment.BlogID)
	}

	BlockHeight = BlockHeight + 10 // Hardcoded value to 100. This can be changed as per requirement.
	
	// Check if the comment is older than the Post. If more than 100 blocks, then return error.
	if comment.CreatedAt > BlockHeight {
		return nil, sdkerrors.Wrapf(types.ErrCommentOld, "Comment created at %d is older than post created at %d", comment.CreatedAt, BlockHeight)
	} else {
		id := k.AppendComment(ctx, comment)
		return &types.MsgCreateCommentResponse{Id: id}, nil
	}
}
```

When Comment's validity is checked, it throws 2 error messages - `ErrID` and `ErrCommendOld`. You can define the error messaged by adding it to errors definition in x/blog/types/errors.go

```go
//...
var (
	ErrCommentOld = sdkerrors.Register(ModuleName, 1300, "")
)

var (
	ErrID = sdkerrors.Register(ModuleName, 1400, "")
)
```

Notice, 2 methods `CreatedAt` and `CreatedId` are being invoked inside `msg_server_create_comment.go`. They are responsible for returning the Block height and latest ID of Post respectively.

Define `CreatedAt` and `CreatedId` method inside `x/blog/keeper/msg_server_create_post.go`

```go
//...

var BlockHeight int64
var BlockId uint64 

func (k msgServer) CreatePost(goCtx context.Context, msg *types.MsgCreatePost) (*types.MsgCreatePostResponse, error) {
  //...
  var post = types.Post{
    //...
    CreatedAt: ctx.BlockHeight(),
  }

    // Define the globally declared `BlockHeight` variable
    BlockHeight = post.CreatedAt

    id := k.AppendPost(ctx, post)
    //...

    // After the post has been appended, fetch the latest id in globally declared `BlockId` variable
    BlockId = id
}

// Define two functions to return BlockId and Height respectively.

func CreatedAt() int64 {
	return BlockHeight
}

func CreatedId() uint64 {
	return BlockId
}
```


## Write Data to the Store

Define the `Comment` type and the `AppendComment` keeper method.

When you define the `Comment` type in a proto file, Starport (with the help of `protoc`) takes care of generating the required Go files.

Create the `proto/blog/comment.proto` file and define the `Comment` message:

```go
syntax = "proto3";
package cosmonaut.blog.blog;
option go_package = "github.com/cosmonaut/blog/x/blog/types";

message Comment {
  string creator = 1;
  uint64 id = 2;
  string title = 3;
  string body = 4; 
  uint64 blogID = 5;
  int64 createdAt = 6;
}
```

The contents of the `comment.proto` file are fairly standard and similar to `post.proto`. The file defines a package name that is used to identify messages, among other things, specifies the Go package where new files are generated, and finally defines `message Comment`. 

Each file save triggers an automatic rebuild.  Now, after you build and start your chain with Starport, the `Comment` type is available.


### Define Keeper Methods

Very similar to implementation done in `x/blog/keeper/post.go` we will implement `x/blog/keeper/comment.go`

To keep a list of comments in what is essentially a key-value store, you need to keep track of the index of the comments you insert. Since both comment values and comment count (index) values are kept in the store, you can use different prefixes: `Comment-value-` and `Comment-count-`. 

Add these prefixes to the `x/blog/types/keys.go` file:

```go
const (
  CommentKey      = "Comment-value-"
  CommentCountKey = "Comment-count-"
)
```

When a `Comment` message is sent to the `AppendComment` function, four actions occur: 

- Get the number of comments in the store (count)
- Add a comment by using the count as an ID
- Increment the count
- Return the count

## Write Data to the Store

Now, inside `x/blog/keeper/comment.go` file, implement `GetCommentCount`, `SetCommentCount` and `AppendCommentCount`

First, implement `GetCommentCount`:

```go
func (k Keeper) GetCommentCount(ctx sdk.Context) uint64 {
    // Get the store using storeKey (which is "blog") and CommentCountKey (which is "Comment-count-")
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentCountKey))
    // Convert the CommentCountKey to bytes
	byteKey := []byte(types.CommentCountKey)
    // Get the value of the count
	bz := store.Get(byteKey)
    // Return zero if the count value is not found (for example, it's the first comment)
	if bz == nil {
		return 0
	}
    // Convert the count into a uint64
	return binary.BigEndian.Uint64(bz)
}
```

Now that `GetCommentCount` returns the correct number of comments in the store, implement `SetCommentCount`:

```go
func (k Keeper) SetCommentCount(ctx sdk.Context, count uint64) {
    // Get the store using storeKey (which is "blog") and CommentCountKey (which is "Comment-count-")
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentCountKey))
    // Convert the CommentCountKey to bytes
	byteKey := []byte(types.CommentCountKey)
    // Convert count from uint64 to string and get bytes
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
    // Set the value of Comment-count- to count
	store.Set(byteKey, bz)
}
```

Now that you have implemented functions for getting the number of comments and setting the comment count, you can implement the logic behind the `AppendComment` function:

```go
package keeper

import (
	"encoding/binary"

	"github.com/cosmonaut/blog/x/blog/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AppendComment(ctx sdk.Context, comment types.Comment) uint64 {
    // Get the current number of comments in the store
	count := k.GetCommentCount(ctx)
    // Assign an ID to the comment based on the number of comments in the store
	comment.Id = count
    // Get the store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentKey))
    // Convert the comment ID into bytes
	byteKey := make([]byte, 8)
	binary.BigEndian.PutUint64(byteKey, comment.Id)
    // Marshal the comment into bytes
	appendedValue := k.cdc.MustMarshal(&comment)
    // Insert the comment bytes using comment ID as a key
	store.Set(byteKey, appendedValue)
    // Update the comment count
	k.SetCommentCount(ctx, count+1)
	return count
}
```

By following these steps, you have implemented all of the code required to create new comments and store them on-chain. Now, when a transaction that contains a message of type `MsgCreateComment` is broadcast, the message is routed to your blog module.

- `x/blog/handler.go` calls `k.CreateComment` which in turn calls `AppendComment`.
- `AppendComment` gets the number of comments from the store, adds a comment using the count as an ID, increments the count, and returns the ID.


## Display Posts

```bash
starport scaffold query comments --response title,body,blogID
```

Very similar to previous blog tutorial, we will make changes to `proto/blog/query.proto`

In the `proto/blog/query.proto` file:

```go
// Import the Comment message
import "blog/comment.proto";

message QueryCommentsRequest {
    // Adding pagination to request
  cosmos.base.query.v1beta1.PageRequest pagination = 1;
}

message QueryPostsResponse {
  // Returning a list of posts
  repeated Post Post = 1;
  // Adding pagination to response
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

message QueryCommentsResponse {
    // Returning a list of comments
  repeated Comment Comment = 1;
    // Adding pagination to response
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```


>>>>> DO from here 

After the types are defined in proto files, you can implement post querying logic. In `grpc_query_comments.go`:

```go
package keeper

import (
	"context"

	"github.com/cosmonaut/blog/x/blog/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Comments(c context.Context, req *types.QueryCommentsRequest) (*types.QueryCommentsResponse, error) {
	// Throw an error if request is nil
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	// Define a variable that will store a list of posts
	var comments []*types.Comment
	// Get context with the information about the environment
	ctx := sdk.UnwrapSDKContext(c)
	// Get the key-value module store using the store key (in our case store key is "chain")
	store := ctx.KVStore(k.storeKey)
	// Get the part of the store that keeps posts (using post key, which is "Post-value-")
	commentStore := prefix.NewStore(store, []byte(types.CommentKey))
	// Paginate the posts store based on PageRequest
	pageRes, err := query.Paginate(commentStore, req.Pagination, func(key []byte, value []byte) error {
		var comment types.Comment
		if err := k.cdc.Unmarshal(value, &comment); err != nil {
			return err
		}
		comments = append(comments, &comment)
		return nil
	})
	// Throw an error if pagination failed
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Return a struct containing a list of posts and pagination info
	return &types.QueryCommentsResponse{Comment: comments, Pagination: pageRes}, nil
}
```

Note: Since we have already gRPC to module handler in previous tutorial, we will not add it again.

## Create Post and Comment

Try it out! If the chain is yet not started, run `starport chain serve`.

Create a post:

```bash
blogd tx blog create-post Uno "This is the first post" --from alice
```

```bash
"body":{"messages":[{"@type":"/cosmonaut.blog.blog.MsgCreatePost","creator":"cosmos1dad8xvsj3dse928r52yayygghwvsggvzlm730p","title":"foo","body":"bar"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}

confirm transaction before signing and broadcasting [y/N]: y
{"height":"6861","txhash":"6086372860704F5F88F4D0A3CF23523CF6DAD2F637E4068B92582E3BB13800DA","codespace":"","code":0,"data":"0A100A0A437265617465506F737412020801","raw_log":"[{\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"CreatePost\"}]}]}]","logs":[{"msg_index":0,"log":"","events":[{"type":"message","attributes":[{"key":"action","value":"CreatePost"}]}]}],"info":"","gas_wanted":"200000","gas_used":"44674","tx":null,"timestamp":""}
```

Create a comment:

```bash
blogd tx blog create-comment 0  Uno "This is the first comment" --from alice
```

```bash
{"body":{"messages":[{"@type":"/cosmonaut.blog.blog.MsgCreateComment","creator":"cosmos17pvwgu36fu37j8y9gc4pasxsj3p26ghmlqvngd","id":"0","title":"Uno","body":"This is the first comment","blogID":"2","createdAt":"0"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}
```
confirm transaction before signing and broadcasting [y/N]: y

```bash
code: 0
codespace: ""
data: 0A270A252F636F736D6F6E6175742E626C6F672E626C6F672E4D7367437265617465436F6D6D656E74
gas_used: "45891"
gas_wanted: "200000"
height: "118"
info: ""
logs:
- events:
  - attributes:
    - key: action
      value: CreateComment
    type: message
  log: ""
  msg_index: 0
raw_log: '[{"events":[{"type":"message","attributes":[{"key":"action","value":"CreateComment"}]}]}]'
timestamp: ""
tx: null
txhash: 0CAFC113D1C73BC0210EFEA5964EBD2EB530311169FB442C5CBF0B5E92521C41
```


## Display Post and Comment

Display post:

```bash
blogd q blog posts
```

```bash
Post:
- body: This is the first post
  createdAt: "38"
  creator: cosmos17pvwgu36fu37j8y9gc4pasxsj3p26ghmlqvngd
  id: "0"
  title: Uno
```

Display comment:

```bash
blogd q blog comments
```

```bash
Comment:
- blogID: "0"
  body: This is the first comment
  createdAt: "98"
  creator: cosmos17pvwgu36fu37j8y9gc4pasxsj3p26ghmlqvngd
  id: "0"
  title: Uno
```

## Edge Cases

1. Add comment to non existent Blog Id 

```bash
blogd tx blog create-comment 53 "Edge1"  "This is the 53 comment" --from alice -y
```

```bash
code: 1400
codespace: blog
data: ""
gas_used: "38151"
gas_wanted: "200000"
height: "1019"
info: ""
logs: []
raw_log: 'failed to execute message; message index: 0: Post Blog Id 53 does not exist for which comment was made: '
timestamp: ""
tx: null
txhash: B99BD295A0B08DF58B9FEC8EB41D467C2F28BD4EC8CDB56FBF30DB728B877ABA
```

2. Add comment to an old Blog post

```bash
blogd tx blog create-comment 0 "Comment" "This is a comment" --from alice -y
```

```bash
code: 1300
codespace: blog
data: ""
gas_used: "38101"
gas_wanted: "200000"
height: "1191"
info: ""
logs: []
raw_log: 'failed to execute message; message index: 0: Comment created at 1191 is older than post created at 1047:'
timestamp: ""
tx: null
txhash: A87AAD5E2E6A26F9B80796D013139E9A18DB286D9CF769BC6AA6601DD64C6A35
```

## Conclusion

Congratulations. You have added comments to blog blockchain! 

You have successfully completed these steps:

* Add comment to existing blog
* Check if the comment is valid
* Use CLI to write and display comment