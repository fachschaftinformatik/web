package auth

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/sid"
	"github.com/go-chi/chi/v5"
)

type PostResponse struct {
	database.GetForumPostRow
	UserVote int64 `json:"user_vote"`
}

// CommentResponse wraps ListForumCommentsRow
type CommentResponse struct {
	database.ListForumCommentsRow
}

type CreatePostRequest struct {
	Title     *string  `json:"title,omitempty"`
	Body      *string  `json:"body,omitempty"`
	Type      *string  `json:"type,omitempty"`
	Pinned    *int64   `json:"pinned,omitempty"`
	Programs  []string `json:"programs,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	EventDate *string  `json:"event_date,omitempty"`
	Location  *string  `json:"location,omitempty"`
	ImageURL  *string  `json:"image_url,omitempty"`
	Links     []string `json:"links,omitempty"`
}

type CommentRequest struct {
	Text     string  `json:"text"`
	ParentID *string `json:"parent_id,omitempty"`
}

type VoteRequest struct {
	Vote int `json:"vote"` // 1 or -1
}

// GetForumPosts lists forum posts
// @Summary List forum posts
// @Tags Forum
// @Param type query string false "Post type (forum, news, event)"
// @Param query query string false "Search query"
// @Param sort query string false "Sort by (new, votes)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} PostResponse
// @Router /forum/posts [get]
func (s *Server) GetForumPosts(w http.ResponseWriter, r *http.Request) {
	_, user, _ := s.authenticate(w, r)

	pType := r.URL.Query().Get("type")
	var typePtr *string
	if pType != "" {
		typePtr = &pType
	}

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	sortBy := r.URL.Query().Get("sort")
	searchQ := r.URL.Query().Get("query")
	var searchPtr *string
	if searchQ != "" {
		searchPtr = &searchQ
	}

	var currentUserID *string
	if user.ID != "" {
		currentUserID = &user.ID
	}

	var posts []database.ListForumPostsRow
	var err error

	if sortBy == "votes" {
		topPosts, err2 := s.DB.ListForumPostsTop(r.Context(), database.ListForumPostsTopParams{
			Type:          typePtr,
			Query:         searchPtr,
			Limit:         limit,
			Offset:        offset,
			CurrentUserID: currentUserID,
		})
		err = err2
		for _, p := range topPosts {
			posts = append(posts, database.ListForumPostsRow(p))
		}
	} else {
		posts, err = s.DB.ListForumPosts(r.Context(), database.ListForumPostsParams{
			Type:          typePtr,
			Query:         searchPtr,
			Limit:         limit,
			Offset:        offset,
			CurrentUserID: currentUserID,
		})
	}

	if err != nil {
		s.Log.Printf("Failed to list forum posts: %v", err)
		s.jsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	totalCount, err := s.DB.CountForumPosts(r.Context(), database.CountForumPostsParams{
		Type:  typePtr,
		Query: searchPtr,
	})
	if err != nil {
		s.Log.Printf("Failed to count forum posts: %v", err)
		totalCount = int64(len(posts))
	}
	w.Header().Set("X-Total-Count", strconv.FormatInt(totalCount, 10))

	resp := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		authorAvatarUrl := *s.formatAvatarURL(&p.AuthorAvatarUrl, p.AuthorID)
		resp = append(resp, PostResponse{
			GetForumPostRow: database.GetForumPostRow{
				ID:              p.ID,
				Title:           p.Title,
				Body:            p.Body,
				AuthorID:        p.AuthorID,
				CreatedAt:       p.CreatedAt,
				UpdatedAt:       p.UpdatedAt,
				Pinned:          p.Pinned,
				Type:            p.Type,
				Programs:        p.Programs,
				Tags:            p.Tags,
				EventDate:       p.EventDate,
				Location:        p.Location,
				ImageUrl:        p.ImageUrl,
				Links:           p.Links,
				AuthorName:      p.AuthorName,
				AuthorAvatarUrl: authorAvatarUrl,
				CommentCount:    p.CommentCount,
				Votes:           p.Votes,
			},
			UserVote: p.UserVote,
		})
	}

	s.respondJSON(w, http.StatusOK, resp)
}

// PostForumPosts creates a new forum post
// @Summary Create forum post
// @Tags Forum
// @Accept json
// @Produce json
// @Param request body CreatePostRequest true "Post Content"
// @Success 201 {object} database.ForumPost
// @Router /forum/posts [post]
func (s *Server) PostForumPosts(w http.ResponseWriter, r *http.Request) {
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	var payload CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if payload.Title == nil || payload.Body == nil {
		s.jsonError(w, "invalid_request_body", "Title and body are required", http.StatusBadRequest)
		return
	}

	pinned := int64(0)
	if payload.Pinned != nil && (user.Role == "admin" || user.Role == "editor") {
		pinned = *payload.Pinned
	}

	programsJSON, _ := json.Marshal(payload.Programs)
	tagsJSON, _ := json.Marshal(payload.Tags)
	linksJSON, _ := json.Marshal(payload.Links)

	post, err := s.DB.CreateForumPost(r.Context(), database.CreateForumPostParams{
		ID:        sid.New(),
		Title:     *payload.Title,
		Body:      *payload.Body,
		AuthorID:  user.ID,
		Type:      payload.Type,
		Pinned:    pinned,
		EventDate: payload.EventDate,
		Location:  payload.Location,
		Programs:  string(programsJSON),
		Tags:      string(tagsJSON),
		ImageUrl:  payload.ImageURL,
		Links:     string(linksJSON),
	})
	if err != nil {
		s.Log.Printf("Failed to create forum post: %v", err)
		s.jsonError(w, "database_error", "Could not create post", http.StatusInternalServerError)
		return
	}

	// Log activity
	activityTargetName := *payload.Title
	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         sid.New(),
		UserID:     user.ID,
		Type:       "POST_CREATED",
		TargetID:   post.ID,
		TargetName: &activityTargetName,
	})

	s.respondJSON(w, http.StatusCreated, post)
}

// GetForumPostsId gets a forum post by ID
// @Summary Get forum post
// @Tags Forum
// @Param id path string true "Post ID"
// @Success 200 {object} PostResponse
// @Router /forum/posts/{id} [get]
func (s *Server) GetForumPostsId(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	_, user, _ := s.authenticate(w, r)
	var currentUserID *string
	if user.ID != "" {
		currentUserID = &user.ID
	}

	post, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{
		ID:            postID,
		CurrentUserID: currentUserID,
	})
	if err != nil {
		s.jsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	resp := PostResponse{
		GetForumPostRow: post,
		UserVote:        post.UserVote,
	}
	resp.AuthorAvatarUrl = *s.formatAvatarURL(&resp.AuthorAvatarUrl, resp.AuthorID)

	s.respondJSON(w, http.StatusOK, resp)
}

// PutForumPostsId updates a forum post
// @Summary Update forum post
// @Tags Forum
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body CreatePostRequest true "Post Content"
// @Success 200 {object} database.ForumPost
// @Router /forum/posts/{id} [put]
func (s *Server) PutForumPostsId(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	existing, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{
		ID:            postID,
		CurrentUserID: nil,
	})
	if err != nil {
		s.jsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	if existing.AuthorID != user.ID && user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "You can only edit your own posts", http.StatusForbidden)
		return
	}

	var payload CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	var pinned *int64
	if payload.Pinned != nil && (user.Role == "admin" || user.Role == "editor") {
		pinned = payload.Pinned
	}

	var programsJSON *string
	if payload.Programs != nil {
		j, _ := json.Marshal(payload.Programs)
		s := string(j)
		programsJSON = &s
	}

	var tagsJSON *string
	if payload.Tags != nil {
		j, _ := json.Marshal(payload.Tags)
		s := string(j)
		tagsJSON = &s
	}

	var linksJSON *string
	if payload.Links != nil {
		j, _ := json.Marshal(payload.Links)
		s := string(j)
		linksJSON = &s
	}

	updated, err := s.DB.UpdateForumPost(r.Context(), database.UpdateForumPostParams{
		ID:        postID,
		Title:     payload.Title,
		Body:      payload.Body,
		Pinned:    pinned,
		EventDate: payload.EventDate,
		Location:  payload.Location,
		ImageUrl:  payload.ImageURL,
		Links:     linksJSON,
		Programs:  programsJSON,
		Tags:      tagsJSON,
	})
	if err != nil {
		s.Log.Printf("Failed to update forum post: %v", err)
		s.jsonError(w, "database_error", "Could not update post", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusOK, updated)
}

// DeleteForumPostsId deletes a forum post
// @Summary Delete forum post
// @Tags Forum
// @Param id path string true "Post ID"
// @Success 204
// @Router /forum/posts/{id} [delete]
func (s *Server) DeleteForumPostsId(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	existing, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{
		ID:            postID,
		CurrentUserID: nil,
	})
	if err != nil {
		s.jsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	if existing.AuthorID != user.ID && user.Role != "admin" && user.Role != "editor" {
		s.jsonError(w, "forbidden", "You can only delete your own posts", http.StatusForbidden)
		return
	}

	if err := s.DB.DeleteForumPost(r.Context(), postID); err != nil {
		s.Log.Printf("Failed to delete forum post: %v", err)
		s.jsonError(w, "database_error", "Could not delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetForumPostsComments lists comments for a post
// @Summary List comments for post
// @Tags Forum
// @Param id path string true "Post ID"
// @Success 200 {array} CommentResponse
// @Router /forum/posts/{id}/comments [get]
func (s *Server) GetForumPostsComments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	_, user, _ := s.authenticate(w, r)
	var currentUserID *string
	if user.ID != "" {
		currentUserID = &user.ID
	}

	comments, err := s.DB.ListForumComments(r.Context(), database.ListForumCommentsParams{
		PostID:        postID,
		CurrentUserID: currentUserID,
	})
	if err != nil {
		s.Log.Printf("Failed to list forum comments: %v", err)
		s.jsonError(w, "database_error", "Could not fetch comments", http.StatusInternalServerError)
		return
	}

	resp := make([]CommentResponse, 0, len(comments))
	for _, c := range comments {
		c.AuthorAvatarUrl = *s.formatAvatarURL(&c.AuthorAvatarUrl, c.AuthorID)
		resp = append(resp, CommentResponse{
			ListForumCommentsRow: c,
		})
	}

	s.respondJSON(w, http.StatusOK, resp)
}

// PostForumPostsComments adds a comment to a post
// @Summary Add comment to post
// @Tags Forum
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body CommentRequest true "Comment Content"
// @Success 201 {object} database.ForumComment
// @Router /forum/posts/{id}/comments [post]
func (s *Server) PostForumPostsComments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	var payload CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	comment, err := s.DB.CreateForumComment(r.Context(), database.CreateForumCommentParams{
		ID:       sid.New(),
		PostID:   postID,
		AuthorID: user.ID,
		ParentID: payload.ParentID,
		Text:     payload.Text,
	})
	if err != nil {
		s.Log.Printf("Failed to create forum comment: %v", err)
		s.jsonError(w, "database_error", "Could not create comment", http.StatusInternalServerError)
		return
	}

	// Log activity
	activityTargetName := "Kommentar zu Beitrag"
	if post, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{ID: postID}); err == nil {
		activityTargetName = post.Title
	}

	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         sid.New(),
		UserID:     user.ID,
		Type:       "COMMENT_ADDED",
		TargetID:   postID, // Log the Post ID, not the Comment ID
		TargetName: &activityTargetName,
	})

	s.respondJSON(w, http.StatusCreated, comment)
}

// PutForumPostsCommentsId updates a comment
// @Summary Update comment
// @Tags Forum
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param request body CommentRequest true "Comment Content"
// @Success 200 {object} database.ForumComment
// @Router /forum/comments/{id} [put]
func (s *Server) PutForumPostsCommentsId(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	var payload CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	// First verify ownership
	existing, err := s.DB.GetForumComment(r.Context(), commentID)
	if err != nil {
		s.jsonError(w, "not_found", "Comment not found", http.StatusNotFound)
		return
	}

	if existing.AuthorID != user.ID && user.Role != "admin" {
		s.jsonError(w, "forbidden", "You can only edit your own comments", http.StatusForbidden)
		return
	}

	updated, err := s.DB.UpdateForumComment(r.Context(), database.UpdateForumCommentParams{
		ID:   commentID,
		Text: payload.Text,
	})
	if err != nil {
		s.Log.Printf("Failed to update comment: %v", err)
		s.jsonError(w, "database_error", "Could not update comment", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusOK, updated)
}

// PostForumCommentsVote votes for a comment
// @Summary Vote for comment
// @Tags Forum
// @Accept json
// @Param id path string true "Comment ID"
// @Param request body VoteRequest true "Vote Value"
// @Success 204
// @Router /forum/comments/{id}/vote [post]
func (s *Server) PostForumCommentsVote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	var payload VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if payload.Vote != 1 && payload.Vote != -1 && payload.Vote != 0 {
		s.jsonError(w, "bad_request", "Invalid vote value", http.StatusBadRequest)
		return
	}

	_, err = s.DB.UpsertForumCommentVote(r.Context(), database.UpsertForumCommentVoteParams{
		CommentID: id,
		UserID:    user.ID,
		Vote:      int64(payload.Vote),
	})
	if err != nil {
		s.Log.Printf("Failed to vote for forum comment: %v", err)
		s.jsonError(w, "database_error", "Could not cast vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostForumPostsVote votes for a post
// @Summary Vote for post
// @Tags Forum
// @Accept json
// @Param id path string true "Post ID"
// @Param request body VoteRequest true "Vote Value"
// @Success 204
// @Router /forum/posts/{id}/vote [post]
func (s *Server) PostForumPostsVote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, user, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	var payload VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if payload.Vote != 1 && payload.Vote != -1 && payload.Vote != 0 {
		s.jsonError(w, "bad_request", "Invalid vote value", http.StatusBadRequest)
		return
	}

	if payload.Vote == 0 {
		// In my simple upsert logic, 0 might not be handled well if I wanted to delete,
		// but I'll just use the upsert for now.
		// Actually, let's just use the Upsert with 0 to clear it if we wanted, or just support 1/-1.
	}

	_, err = s.DB.UpsertForumVote(r.Context(), database.UpsertForumVoteParams{
		PostID: id,
		UserID: user.ID,
		Vote:   int64(payload.Vote),
	})
	if err != nil {
		s.Log.Printf("Failed to vote for forum post: %v", err)
		s.jsonError(w, "database_error", "Could not cast vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
