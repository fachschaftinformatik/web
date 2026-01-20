package handler

import (
	"encoding/json"
	"github.com/fachschaftinformatik/web/internal/id"
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/go-chi/chi/v5"
)

// @Summary List forum posts
// @Tags Forum
// @Param type query string false "Post type (forum, news, event)"
// @Param query query string false "Search query"
// @Param sort query string false "Sort by (new, votes)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.PostResponse
// @Router /forum/posts [get]
func (s *Server) GetForumPosts(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

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
		s.Log.Error("Failed to list forum posts", "err", err)
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	totalCount, err := s.DB.CountForumPosts(r.Context(), database.CountForumPostsParams{
		Type:  typePtr,
		Query: searchPtr,
	})
	if err != nil {
		s.Log.Error("Failed to count forum posts", "err", err)
		totalCount = int64(len(posts))
	}
	w.Header().Set("X-Total-Count", strconv.FormatInt(totalCount, 10))

	resp := make([]dto.PostResponse, 0, len(posts))
	for _, p := range posts {
		var programs []string
		_ = json.Unmarshal([]byte(p.Programs), &programs)
		var tags []string
		_ = json.Unmarshal([]byte(p.Tags), &tags)
		var links []string
		_ = json.Unmarshal([]byte(p.Links), &links)

		resp = append(resp, dto.PostResponse{
			ID:              p.ID,
			Title:           p.Title,
			Body:            p.Body,
			AuthorID:        p.AuthorID,
			AuthorName:      p.AuthorName,
			AuthorAvatarUrl: *s.FormatAvatarURL(&p.AuthorAvatarUrl, p.AuthorID, 0), // Already masked in query if private
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
			Pinned:          p.Pinned,
			Type:            p.Type,
			Programs:        programs,
			Tags:            tags,
			EventDate:       p.EventDate,
			Location:        p.Location,
			ImageUrl:        p.ImageUrl,
			Links:           links,
			CommentCount:    p.CommentCount,
			Votes:           p.Votes,
			UserVote:        p.UserVote,
			Active:          p.Active,
		})
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Create forum post
// @Tags Forum
// @Accept json
// @Produce json
// @Param request body dto.CreatePostRequest true "Post Content"
// @Success 201 {object} dto.PostResponse
// @Router /forum/posts [post]
func (s *Server) PostForumPosts(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	var payload dto.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	postType := "forum"
	if payload.Type != nil {
		postType = *payload.Type
	}

	pinned := int64(0)
	if payload.Pinned != nil && (user.Role == "admin" || user.Role == "editor") {
		pinned = *payload.Pinned
	}

	programsJSON, _ := json.Marshal(payload.Programs)
	tagsJSON, _ := json.Marshal(payload.Tags)
	linksJSON, _ := json.Marshal(payload.Links)

	post, err := s.DB.CreateForumPost(r.Context(), database.CreateForumPostParams{
		ID:        id.New(),
		Title:     s.Sanitize(*payload.Title),
		Body:      s.Sanitize(*payload.Body),
		AuthorID:  user.ID,
		Type:      postType,
		Pinned:    pinned,
		EventDate: payload.EventDate,
		Location:  payload.Location,
		Programs:  string(programsJSON),
		Tags:      string(tagsJSON),
		ImageUrl:  payload.ImageURL,
		Links:     string(linksJSON),
	})
	if err != nil {
		s.Log.Error("Failed to create forum post", "err", err)
		s.JsonError(w, "database_error", "Could not create post", http.StatusInternalServerError)
		return
	}

	activityTargetName := post.Title
	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         id.New(),
		UserID:     user.ID,
		Type:       "POST_CREATED",
		TargetID:   post.ID,
		TargetName: &activityTargetName,
	})

	s.RespondJSON(w, http.StatusCreated, s.toPostResponse(post, user))
}

// @Summary Get forum post
// @Tags Forum
// @Param id path string true "Post ID"
// @Success 200 {object} dto.PostResponse
// @Router /forum/posts/{id} [get]
func (s *Server) GetForumPostsId(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	user, _ := s.User(r)
	var currentUserID *string
	if user.ID != "" {
		currentUserID = &user.ID
	}

	post, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{
		ID:            postID,
		CurrentUserID: currentUserID,
	})
	if err != nil {
		s.JsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	var programs []string
	_ = json.Unmarshal([]byte(post.Programs), &programs)
	var tags []string
	_ = json.Unmarshal([]byte(post.Tags), &tags)
	var links []string
	_ = json.Unmarshal([]byte(post.Links), &links)

	resp := dto.PostResponse{
		ID:              post.ID,
		Title:           post.Title,
		Body:            post.Body,
		AuthorID:        post.AuthorID,
		AuthorName:      post.AuthorName,
		AuthorAvatarUrl: *s.FormatAvatarURL(&post.AuthorAvatarUrl, post.AuthorID, 0), // Masked in query
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
		Pinned:          post.Pinned,
		Type:            post.Type,
		Programs:        programs,
		Tags:            tags,
		EventDate:       post.EventDate,
		Location:        post.Location,
		ImageUrl:        post.ImageUrl,
		Links:           links,
		CommentCount:    post.CommentCount,
		Votes:           post.Votes,
		UserVote:        post.UserVote,
		Active:          post.Active,
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Update forum post
// @Tags Forum
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body dto.CreatePostRequest true "Post Content"
// @Success 200 {object} dto.PostResponse
// @Router /forum/posts/{id} [put]
func (s *Server) PutForumPostsId(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	user, _ := s.User(r)

	existing, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{
		ID:            postID,
		CurrentUserID: nil,
	})
	if err != nil {
		s.JsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	if existing.AuthorID != user.ID && user.Role != "admin" && user.Role != "editor" {
		s.JsonError(w, "forbidden", "You can only edit your own posts", http.StatusForbidden)
		return
	}

	var payload dto.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	var pinned *int64
	if payload.Pinned != nil && (user.Role == "admin" || user.Role == "editor") {
		pinned = payload.Pinned
	}

	var programsJSON *string
	if payload.Programs != nil {
		j, _ := json.Marshal(payload.Programs)
		str := string(j)
		programsJSON = &str
	}

	var tagsJSON *string
	if payload.Tags != nil {
		j, _ := json.Marshal(payload.Tags)
		str := string(j)
		tagsJSON = &str
	}

	var linksJSON *string
	if payload.Links != nil {
		j, _ := json.Marshal(payload.Links)
		str := string(j)
		linksJSON = &str
	}

	var titlePtr, bodyPtr *string
	if payload.Title != nil {
		t := s.Sanitize(*payload.Title)
		titlePtr = &t
	}
	if payload.Body != nil {
		b := s.Sanitize(*payload.Body)
		bodyPtr = &b
	}

	updated, err := s.DB.UpdateForumPost(r.Context(), database.UpdateForumPostParams{
		ID:        postID,
		Title:     titlePtr,
		Body:      bodyPtr,
		Pinned:    pinned,
		EventDate: payload.EventDate,
		Location:  payload.Location,
		ImageUrl:  payload.ImageURL,
		Links:     linksJSON,
		Programs:  programsJSON,
		Tags:      tagsJSON,
	})
	if err != nil {
		s.Log.Error("Failed to update forum post", "err", err)
		s.JsonError(w, "database_error", "Could not update post", http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusOK, s.toPostResponse(updated, user))
}

// @Summary Delete forum post
// @Tags Forum
// @Param id path string true "Post ID"
// @Success 204
// @Router /forum/posts/{id} [delete]
func (s *Server) DeleteForumPostsId(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	user, _ := s.User(r)

	existing, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{
		ID:            postID,
		CurrentUserID: nil,
	})
	if err != nil {
		s.JsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	if existing.AuthorID != user.ID && user.Role != "admin" && user.Role != "editor" {
		s.JsonError(w, "forbidden", "You can only delete your own posts", http.StatusForbidden)
		return
	}

	if err := s.DB.DeleteForumPost(r.Context(), postID); err != nil {
		s.Log.Error("Failed to delete forum post", "err", err)
		s.JsonError(w, "database_error", "Could not delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary List comments for post
// @Tags Forum
// @Param id path string true "Post ID"
// @Success 200 {array} dto.CommentResponse
// @Router /forum/posts/{id}/comments [get]
func (s *Server) GetForumPostsComments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	user, _ := s.User(r)
	var currentUserID *string
	if user.ID != "" {
		currentUserID = &user.ID
	}

	comments, err := s.DB.ListForumComments(r.Context(), database.ListForumCommentsParams{
		PostID:        postID,
		CurrentUserID: currentUserID,
	})
	if err != nil {
		s.Log.Error("Failed to list forum comments", "err", err)
		s.JsonError(w, "database_error", "Could not fetch comments", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.CommentResponse, 0, len(comments))
	for _, c := range comments {
		resp = append(resp, dto.CommentResponse{
			ID:              c.ID,
			PostID:          c.PostID,
			AuthorID:        c.AuthorID,
			AuthorName:      c.AuthorName,
			AuthorAvatarUrl: *s.FormatAvatarURL(&c.AuthorAvatarUrl, c.AuthorID, 0), // Masked in query
			ParentID:        c.ParentID,
			Text:            c.Text,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
			Votes:           c.Votes,
			UserVote:        c.UserVote,
		})
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Add comment to post
// @Tags Forum
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body dto.CommentRequest true "Comment Content"
// @Success 201 {object} dto.CommentResponse
// @Router /forum/posts/{id}/comments [post]
func (s *Server) PostForumPostsComments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	user, _ := s.User(r)

	var payload dto.CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	comment, err := s.DB.CreateForumComment(r.Context(), database.CreateForumCommentParams{
		ID:       id.New(),
		PostID:   postID,
		AuthorID: user.ID,
		ParentID: payload.ParentID,
		Text:     s.Sanitize(payload.Text),
	})
	if err != nil {
		s.Log.Error("Failed to create forum comment", "err", err)
		s.JsonError(w, "database_error", "Could not create comment", http.StatusInternalServerError)
		return
	}

	activityTargetName := "Kommentar zu Beitrag"
	if post, err := s.DB.GetForumPost(r.Context(), database.GetForumPostParams{ID: postID}); err == nil {
		activityTargetName = post.Title
	}

	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         id.New(),
		UserID:     user.ID,
		Type:       "COMMENT_ADDED",
		TargetID:   postID,
		TargetName: &activityTargetName,
	})

	s.RespondJSON(w, http.StatusCreated, s.toCommentResponse(comment, user))
}

// @Summary Update comment
// @Tags Forum
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param request body dto.CommentRequest true "Comment Content"
// @Success 200 {object} dto.CommentResponse
// @Router /forum/comments/{id} [put]
func (s *Server) PutForumPostsCommentsId(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	user, _ := s.User(r)

	var payload dto.CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	existing, err := s.DB.GetForumComment(r.Context(), commentID)
	if err != nil {
		s.JsonError(w, "not_found", "Comment not found", http.StatusNotFound)
		return
	}

	if existing.AuthorID != user.ID && user.Role != "admin" {
		s.JsonError(w, "forbidden", "You can only edit your own comments", http.StatusForbidden)
		return
	}

	updated, err := s.DB.UpdateForumComment(r.Context(), database.UpdateForumCommentParams{
		ID:   commentID,
		Text: s.Sanitize(payload.Text),
	})
	if err != nil {
		s.Log.Error("Failed to update comment", "err", err)
		s.JsonError(w, "database_error", "Could not update comment", http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusOK, s.toCommentResponse(updated, user))
}

// @Summary Vote for comment
// @Tags Forum
// @Accept json
// @Param id path string true "Comment ID"
// @Param request body dto.VoteRequest true "Vote Value"
// @Success 204
// @Router /forum/comments/{id}/vote [post]
func (s *Server) PostForumCommentsVote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, _ := s.User(r)

	var payload dto.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	_, err := s.DB.UpsertForumCommentVote(r.Context(), database.UpsertForumCommentVoteParams{
		CommentID: id,
		UserID:    user.ID,
		Vote:      int64(payload.Vote),
	})
	if err != nil {
		s.Log.Error("Failed to vote for forum comment", "err", err)
		s.JsonError(w, "database_error", "Could not cast vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Vote for post
// @Tags Forum
// @Accept json
// @Param id path string true "Post ID"
// @Param request body dto.VoteRequest true "Vote Value"
// @Success 204
// @Router /forum/posts/{id}/vote [post]
func (s *Server) PostForumPostsVote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, _ := s.User(r)

	var payload dto.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	_, err := s.DB.UpsertForumVote(r.Context(), database.UpsertForumVoteParams{
		PostID: id,
		UserID: user.ID,
		Vote:   int64(payload.Vote),
	})
	if err != nil {
		s.Log.Error("Failed to vote for forum post", "err", err)
		s.JsonError(w, "database_error", "Could not cast vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) toPostResponse(p database.ForumPost, author database.User) dto.PostResponse {
	var programs []string
	_ = json.Unmarshal([]byte(p.Programs), &programs)
	var tags []string
	_ = json.Unmarshal([]byte(p.Tags), &tags)
	var links []string
	_ = json.Unmarshal([]byte(p.Links), &links)

	authorName := author.Name
	if author.Private == 1 {
		authorName = "Anonym"
	}

	return dto.PostResponse{
		ID:              p.ID,
		Title:           p.Title,
		Body:            p.Body,
		AuthorID:        p.AuthorID,
		AuthorName:      authorName,
		AuthorAvatarUrl: *s.FormatAvatarURL(author.AvatarUrl, author.ID, author.Private),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Pinned:          p.Pinned,
		Type:            p.Type,
		Programs:        programs,
		Tags:            tags,
		EventDate:       p.EventDate,
		Location:        p.Location,
		ImageUrl:        p.ImageUrl,
		Links:           links,
		Active:          p.Active,
	}
}

func (s *Server) toCommentResponse(c database.ForumComment, author database.User) dto.CommentResponse {
	authorName := author.Name
	if author.Private == 1 {
		authorName = "Anonym"
	}

	return dto.CommentResponse{
		ID:              c.ID,
		PostID:          c.PostID,
		AuthorID:        c.AuthorID,
		AuthorName:      authorName,
		AuthorAvatarUrl: *s.FormatAvatarURL(author.AvatarUrl, author.ID, author.Private),
		ParentID:        c.ParentID,
		Text:            c.Text,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
