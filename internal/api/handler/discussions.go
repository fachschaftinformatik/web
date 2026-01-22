package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/go-chi/chi/v5"
)

// @Summary List discussion posts
// @Tags Discussions
// @Param type query string false "Post type (discussion, news, event)"
// @Param program_id query string false "Program ID filter"
// @Param query query string false "Search query"
// @Param sort query string false "Sort by (new, votes)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.DiscussionPostResponse
// @Router /discussions [get]
func (s *Server) GetDiscussions(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	pType := r.URL.Query().Get("type")
	var typePtr *string
	if pType != "" {
		typePtr = &pType
	}

	pID := r.URL.Query().Get("program_id")
	var programIDPtr *int64
	if pID != "" {
		if idVal, err := id.Parse(pID); err == nil {
			v := int64(idVal)
			programIDPtr = &v
		}
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

	var currentUserID *int64
	if user.ID != 0 {
		cid := int64(user.ID)
		currentUserID = &cid
	}

	var postsRows []database.ListDiscussionPostsRow
	var err error

	if sortBy == "votes" {
		topPosts, err2 := s.DB.ListDiscussionPostsTop(r.Context(), database.ListDiscussionPostsTopParams{
			Type:          typePtr,
			Query:         searchPtr,
			Limit:         limit,
			Offset:        offset,
			CurrentUserID: currentUserID,
			ProgramID:     programIDPtr,
		})
		err = err2
		for _, p := range topPosts {
			postsRows = append(postsRows, database.ListDiscussionPostsRow{
				ID:            p.ID,
				Title:         p.Title,
				Body:          p.Body,
				UserID:        p.UserID,
				Pinned:        p.Pinned,
				Type:          p.Type,
				EventDate:     p.EventDate,
				Location:      p.Location,
				ImageUrl:      p.ImageUrl,
				CreatedAt:     p.CreatedAt,
				UpdatedAt:     p.UpdatedAt,
				DeletedAt:     p.DeletedAt,
				UserName:      p.UserName,
				UserAvatarUrl: p.UserAvatarUrl,
				CommentCount:  p.CommentCount,
				Votes:         p.Votes,
				UserVote:      p.UserVote,
			})
		}
	} else {
		postsRows, err = s.DB.ListDiscussionPosts(r.Context(), database.ListDiscussionPostsParams{
			Type:          typePtr,
			Query:         searchPtr,
			Limit:         limit,
			Offset:        offset,
			CurrentUserID: currentUserID,
			ProgramID:     programIDPtr,
		})
	}

	if err != nil {
		s.Log.Error("Failed to list discussion posts", "err", err)
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	totalCount, err := s.DB.CountDiscussionPosts(r.Context(), database.CountDiscussionPostsParams{
		Type:      typePtr,
		Query:     searchPtr,
		ProgramID: programIDPtr,
	})
	if err != nil {
		s.Log.Error("Failed to count discussion posts", "err", err)
		totalCount = int64(len(postsRows))
	}
	w.Header().Set("X-Total-Count", strconv.FormatInt(totalCount, 10))

	resp := make([]dto.DiscussionPostResponse, 0, len(postsRows))
	for _, p := range postsRows {
		programs, _ := s.DB.GetPostPrograms(r.Context(), p.ID)
		tags, _ := s.DB.GetPostTags(r.Context(), p.ID)
		links, _ := s.DB.GetPostLinks(r.Context(), p.ID)

		dtoPrograms := make([]dto.PostProgramResponse, len(programs))
		for i, prog := range programs {
			dtoPrograms[i] = dto.PostProgramResponse{
				ID:   id.ID(prog.ID),
				Name: prog.Name,
			}
		}

		dtoLinks := make([]dto.Link, len(links))
		for i, l := range links {
			label := ""
			if l.Label != nil {
				label = *l.Label
			}
			dtoLinks[i] = dto.Link{URL: l.Url, Label: label}
		}

		resp = append(resp, dto.DiscussionPostResponse{
			ID:            id.ID(p.ID),
			Title:         p.Title,
			Body:          p.Body,
			UserID:        id.ID(p.UserID),
			UserName:      p.UserName,
			UserAvatarUrl: *s.FormatAvatarURL(&p.UserAvatarUrl, id.ID(p.UserID), 0),
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
			Pinned:        p.Pinned,
			Type:          p.Type,
			Programs:      dtoPrograms,
			Tags:          tags,
			EventDate:     p.EventDate,
			Location:      p.Location,
			ImageUrl:      p.ImageUrl,
			Links:         dtoLinks,
			CommentCount:  p.CommentCount,
			Votes:         p.Votes,
			UserVote:      p.UserVote,
		})
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Create discussion post
// @Tags Discussions
// @Accept json
// @Produce json
// @Param request body dto.CreateDiscussionPostRequest true "Post Content"
// @Success 201 {object} dto.DiscussionPostResponse
// @Router /discussions [post]
func (s *Server) PostDiscussions(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	var payload dto.CreateDiscussionPostRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	postType := "discussion"
	if payload.Type != "" {
		postType = payload.Type
	}

	pinned := int64(0)
	if payload.Pinned && (user.Role == "admin" || user.Role == "editor") {
		pinned = 1
	}

	post, err := s.DB.CreateDiscussionPost(r.Context(), database.CreateDiscussionPostParams{
		ID:        int64(id.New()),
		Title:     s.Sanitize(payload.Title),
		Body:      s.Sanitize(payload.Body),
		UserID:    user.ID,
		Type:      postType,
		Pinned:    pinned,
		EventDate: payload.EventDate,
		Location:  payload.Location,
		ImageUrl:  payload.ImageURL,
	})
	if err != nil {
		s.Log.Error("Failed to create discussion post", "err", err)
		s.JsonError(w, "database_error", "Could not create post", http.StatusInternalServerError)
		return
	}

	for _, pid := range payload.Programs {
		_ = s.DB.AddProgramToPost(r.Context(), database.AddProgramToPostParams{PostID: post.ID, ProgramID: int64(pid)})
	}
	for _, tag := range payload.Tags {
		_ = s.DB.AddTagToPost(r.Context(), database.AddTagToPostParams{PostID: post.ID, Tag: tag})
	}
	for _, l := range payload.Links {
		var label *string
		if l.Label != "" {
			label = &l.Label
		}
		_, _ = s.DB.AddLinkToPost(r.Context(), database.AddLinkToPostParams{PostID: post.ID, Url: l.URL, Label: label})
	}

	activityTargetName := post.Title
	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         int64(id.New()),
		UserID:     user.ID,
		Type:       "POST_CREATED",
		TargetID:   strconv.FormatInt(post.ID, 10),
		TargetName: &activityTargetName,
	})

	link := "/discussions/" + id.ID(post.ID).String()
	s.broadcastNotification(r, "Neuer Beitrag", post.Title, "discussion", link)

	postPrograms, _ := s.DB.GetPostPrograms(r.Context(), post.ID)
	dtoPostPrograms := make([]dto.PostProgramResponse, len(postPrograms))
	for i, p := range postPrograms {
		dtoPostPrograms[i] = dto.PostProgramResponse{ID: id.ID(p.ID), Name: p.Name}
	}

	s.RespondJSON(w, http.StatusCreated, s.toDiscussionPostResponse(post, user, dtoPostPrograms, payload.Tags, payload.Links))
}

// @Summary Get discussion post
// @Tags Discussions
// @Param id path string true "Post ID"
// @Success 200 {object} dto.DiscussionPostResponse
// @Router /discussions/{postId} [get]
func (s *Server) GetDiscussionsId(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postId")
	pid, err := id.Parse(postIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid post ID", http.StatusBadRequest)
		return
	}

	user, _ := s.User(r)
	var currentUserID *int64
	if user.ID != 0 {
		cid := int64(user.ID)
		currentUserID = &cid
	}

	post, err := s.DB.GetDiscussionPost(r.Context(), database.GetDiscussionPostParams{
		ID:            int64(pid),
		CurrentUserID: currentUserID,
	})
	if err != nil {
		s.JsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	programs, _ := s.DB.GetPostPrograms(r.Context(), post.ID)
	tags, _ := s.DB.GetPostTags(r.Context(), post.ID)
	links, _ := s.DB.GetPostLinks(r.Context(), post.ID)

	dtoPrograms := make([]dto.PostProgramResponse, len(programs))
	for i, prog := range programs {
		dtoPrograms[i] = dto.PostProgramResponse{
			ID:   id.ID(prog.ID),
			Name: prog.Name,
		}
	}

	dtoLinks := make([]dto.Link, len(links))
	for i, l := range links {
		label := ""
		if l.Label != nil {
			label = *l.Label
		}
		dtoLinks[i] = dto.Link{URL: l.Url, Label: label}
	}

	resp := dto.DiscussionPostResponse{
		ID:            id.ID(post.ID),
		Title:         post.Title,
		Body:          post.Body,
		UserID:        id.ID(post.UserID),
		UserName:      post.UserName,
		UserAvatarUrl: *s.FormatAvatarURL(&post.UserAvatarUrl, id.ID(post.UserID), 0),
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		Pinned:        post.Pinned,
		Type:          post.Type,
		Programs:      dtoPrograms,
		Tags:          tags,
		EventDate:     post.EventDate,
		Location:      post.Location,
		ImageUrl:      post.ImageUrl,
		Links:         dtoLinks,
		CommentCount:  post.CommentCount,
		Votes:         post.Votes,
		UserVote:      post.UserVote,
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Update discussion post
// @Tags Discussions
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body dto.CreateDiscussionPostRequest true "Post Content"
// @Success 200 {object} dto.DiscussionPostResponse
// @Router /discussions/{postId} [put]
func (s *Server) PutDiscussionsId(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postId")
	pid, err := id.Parse(postIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid post ID", http.StatusBadRequest)
		return
	}

	user, _ := s.User(r)

	existing, err := s.DB.GetDiscussionPost(r.Context(), database.GetDiscussionPostParams{
		ID:            int64(pid),
		CurrentUserID: nil,
	})
	if err != nil {
		s.JsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	if id.ID(existing.UserID) != id.ID(user.ID) && user.Role != "admin" && user.Role != "editor" {
		s.JsonError(w, "forbidden", "You can only edit your own posts", http.StatusForbidden)
		return
	}

	var payload dto.CreateDiscussionPostRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	pinned := existing.Pinned
	if user.Role == "admin" || user.Role == "editor" {
		if payload.Pinned {
			pinned = 1
		} else {
			pinned = 0
		}
	}

	updated, err := s.DB.UpdateDiscussionPost(r.Context(), database.UpdateDiscussionPostParams{
		ID:        int64(pid),
		Title:     &payload.Title,
		Body:      &payload.Body,
		Pinned:    &pinned,
		EventDate: payload.EventDate,
		Location:  payload.Location,
		ImageUrl:  payload.ImageURL,
	})
	if err != nil {
		s.Log.Error("Failed to update discussion post", "err", err)
		s.JsonError(w, "database_error", "Could not update post", http.StatusInternalServerError)
		return
	}

	_ = s.DB.ClearPostPrograms(r.Context(), updated.ID)
	_ = s.DB.ClearPostTags(r.Context(), updated.ID)
	_ = s.DB.ClearPostLinks(r.Context(), updated.ID)

	for _, pid := range payload.Programs {
		_ = s.DB.AddProgramToPost(r.Context(), database.AddProgramToPostParams{PostID: updated.ID, ProgramID: int64(pid)})
	}
	for _, tag := range payload.Tags {
		_ = s.DB.AddTagToPost(r.Context(), database.AddTagToPostParams{PostID: updated.ID, Tag: tag})
	}
	for _, l := range payload.Links {
		var label *string
		if l.Label != "" {
			label = &l.Label
		}
		_, _ = s.DB.AddLinkToPost(r.Context(), database.AddLinkToPostParams{PostID: updated.ID, Url: l.URL, Label: label})
	}

	postPrograms, _ := s.DB.GetPostPrograms(r.Context(), updated.ID)
	dtoPostPrograms := make([]dto.PostProgramResponse, len(postPrograms))
	for i, p := range postPrograms {
		dtoPostPrograms[i] = dto.PostProgramResponse{ID: id.ID(p.ID), Name: p.Name}
	}

	s.RespondJSON(w, http.StatusOK, s.toDiscussionPostResponse(updated, user, dtoPostPrograms, payload.Tags, payload.Links))
}

// @Summary Delete discussion post
// @Tags Discussions
// @Param id path string true "Post ID"
// @Success 204
// @Router /discussions/{postId} [delete]
func (s *Server) DeleteDiscussionsId(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postId")
	pid, err := id.Parse(postIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid post ID", http.StatusBadRequest)
		return
	}

	user, _ := s.User(r)

	existing, err := s.DB.GetDiscussionPost(r.Context(), database.GetDiscussionPostParams{
		ID:            int64(pid),
		CurrentUserID: nil,
	})
	if err != nil {
		s.JsonError(w, "not_found", "Post not found", http.StatusNotFound)
		return
	}

	if id.ID(existing.UserID) != id.ID(user.ID) && user.Role != "admin" && user.Role != "editor" {
		s.JsonError(w, "forbidden", "You can only delete your own posts", http.StatusForbidden)
		return
	}

	if err := s.DB.DeleteDiscussionPost(r.Context(), int64(pid)); err != nil {
		s.Log.Error("Failed to delete discussion post", "err", err)
		s.JsonError(w, "database_error", "Could not delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary List comments for post
// @Tags Discussions
// @Param id path string true "Post ID"
// @Success 200 {array} dto.DiscussionCommentResponse
// @Router /discussions/{postId}/comments [get]
func (s *Server) GetDiscussionsComments(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postId")
	pid, err := id.Parse(postIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid post ID", http.StatusBadRequest)
		return
	}

	user, _ := s.User(r)
	var currentUserID *int64
	if user.ID != 0 {
		cid := int64(user.ID)
		currentUserID = &cid
	}

	comments, err := s.DB.ListDiscussionComments(r.Context(), database.ListDiscussionCommentsParams{
		PostID:        int64(pid),
		CurrentUserID: currentUserID,
	})
	if err != nil {
		s.Log.Error("Failed to list discussion comments", "err", err)
		s.JsonError(w, "database_error", "Could not fetch comments", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.DiscussionCommentResponse, 0, len(comments))
	for _, c := range comments {
		resp = append(resp, dto.DiscussionCommentResponse{
			ID:            id.ID(c.ID),
			PostID:        id.ID(c.PostID),
			UserID:        id.ID(c.UserID),
			UserName:      c.UserName,
			UserAvatarUrl: *s.FormatAvatarURL(&c.UserAvatarUrl, id.ID(c.UserID), 0),
			ParentID:      (*id.ID)(c.ParentID),
			Text:          c.Text,
			CreatedAt:     c.CreatedAt,
			UpdatedAt:     c.UpdatedAt,
			Votes:         c.Votes,
			UserVote:      c.UserVote,
		})
	}

	s.RespondJSON(w, http.StatusOK, resp)
}

// @Summary Add comment to post
// @Tags Discussions
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body dto.DiscussionCommentRequest true "Comment Content"
// @Success 201 {object} dto.DiscussionCommentResponse
// @Router /discussions/{postId}/comments [post]
func (s *Server) PostDiscussionsComments(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postId")
	pid, err := id.Parse(postIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid post ID", http.StatusBadRequest)
		return
	}

	user, _ := s.User(r)

	var payload dto.DiscussionCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	comment, err := s.DB.CreateDiscussionComment(r.Context(), database.CreateDiscussionCommentParams{
		ID:       int64(id.New()),
		PostID:   int64(pid),
		UserID:   user.ID,
		ParentID: (*int64)(payload.ParentID),
		Text:     s.Sanitize(payload.Text),
	})
	if err != nil {
		s.Log.Error("Failed to create discussion comment", "err", err)
		s.JsonError(w, "database_error", "Could not create comment", http.StatusInternalServerError)
		return
	}

	activityTargetName := "Kommentar zu Beitrag"
	if post, err := s.DB.GetDiscussionPost(r.Context(), database.GetDiscussionPostParams{ID: int64(pid)}); err == nil {
		activityTargetName = post.Title
	}

	_, _ = s.DB.CreateActivity(r.Context(), database.CreateActivityParams{
		ID:         int64(id.New()),
		UserID:     user.ID,
		Type:       "COMMENT_ADDED",
		TargetID:   strconv.FormatInt(int64(pid), 10),
		TargetName: &activityTargetName,
	})

	link := "/discussions/" + id.ID(pid).String()
	s.broadcastNotification(r, "Neuer Kommentar", activityTargetName, "discussion", link)

	s.RespondJSON(w, http.StatusCreated, s.toDiscussionCommentResponse(comment, user))
}

// @Summary Update comment
// @Tags Discussions
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param request body dto.DiscussionCommentRequest true "Comment Content"
// @Success 200 {object} dto.DiscussionCommentResponse
// @Router /discussions/comments/{commentId} [put]
func (s *Server) PutDiscussionsCommentsId(w http.ResponseWriter, r *http.Request) {
	commentIDStr := chi.URLParam(r, "commentId")
	cid, err := id.Parse(commentIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid comment ID", http.StatusBadRequest)
		return
	}

	user, _ := s.User(r)

	var payload dto.DiscussionCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	existing, err := s.DB.GetDiscussionComment(r.Context(), int64(cid))
	if err != nil {
		s.JsonError(w, "not_found", "Comment not found", http.StatusNotFound)
		return
	}

	if id.ID(existing.UserID) != id.ID(user.ID) && user.Role != "admin" {
		s.JsonError(w, "forbidden", "You can only edit your own comments", http.StatusForbidden)
		return
	}

	updated, err := s.DB.UpdateDiscussionComment(r.Context(), database.UpdateDiscussionCommentParams{
		ID:   int64(cid),
		Text: s.Sanitize(payload.Text),
	})
	if err != nil {
		s.Log.Error("Failed to update discussion comment", "err", err)
		s.JsonError(w, "database_error", "Could not update comment", http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusOK, s.toDiscussionCommentResponse(updated, user))
}

// @Summary Vote for comment
// @Tags Discussions
// @Accept json
// @Param id path string true "Comment ID"
// @Param request body dto.VoteRequest true "Vote Value"
// @Success 204
// @Router /discussions/comments/{commentId}/vote [post]
func (s *Server) PostDiscussionsCommentsVote(w http.ResponseWriter, r *http.Request) {
	commentIDStr := chi.URLParam(r, "commentId")
	cid, err := id.Parse(commentIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid comment ID", http.StatusBadRequest)
		return
	}

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

	_, err = s.DB.UpsertDiscussionCommentVote(r.Context(), database.UpsertDiscussionCommentVoteParams{
		CommentID: int64(cid),
		UserID:    user.ID,
		Vote:      int64(payload.Vote),
	})
	if err != nil {
		s.Log.Error("Failed to vote for discussion comment", "err", err)
		s.JsonError(w, "database_error", "Could not cast vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Vote for post
// @Tags Discussions
// @Accept json
// @Param id path string true "Post ID"
// @Param request body dto.VoteRequest true "Vote Value"
// @Success 204
// @Router /discussions/{postId}/vote [post]
func (s *Server) PostDiscussionsVote(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postId")
	pid, err := id.Parse(postIDStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid post ID", http.StatusBadRequest)
		return
	}

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

	_, err = s.DB.UpsertDiscussionVote(r.Context(), database.UpsertDiscussionVoteParams{
		PostID: int64(pid),
		UserID: user.ID,
		Vote:   int64(payload.Vote),
	})
	if err != nil {
		s.Log.Error("Failed to vote for discussion post", "err", err)
		s.JsonError(w, "database_error", "Could not cast vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) toDiscussionPostResponse(p database.DiscussionPost, author database.User, programs []dto.PostProgramResponse, tags []string, links []dto.Link) dto.DiscussionPostResponse {
	authorName := author.Name
	if author.Private == 1 {
		authorName = "Anonym"
	}

	dtoLinks := make([]dto.Link, len(links))
	copy(dtoLinks, links)

	return dto.DiscussionPostResponse{
		ID:            id.ID(p.ID),
		Title:         p.Title,
		Body:          p.Body,
		UserID:        id.ID(p.UserID),
		UserName:      authorName,
		UserAvatarUrl: *s.FormatAvatarURL(author.AvatarUrl, id.ID(author.ID), author.Private),
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Pinned:        p.Pinned,
		Type:          p.Type,
		Programs:      programs,
		Tags:          tags,
		EventDate:     p.EventDate,
		Location:      p.Location,
		ImageUrl:      p.ImageUrl,
		Links:         dtoLinks,
	}
}

func (s *Server) toDiscussionCommentResponse(c database.DiscussionComment, author database.User) dto.DiscussionCommentResponse {
	authorName := author.Name
	if author.Private == 1 {
		authorName = "Anonym"
	}

	return dto.DiscussionCommentResponse{
		ID:            id.ID(c.ID),
		PostID:        id.ID(c.PostID),
		UserID:        id.ID(c.UserID),
		UserName:      authorName,
		UserAvatarUrl: *s.FormatAvatarURL(author.AvatarUrl, id.ID(author.ID), author.Private),
		ParentID:      (*id.ID)(c.ParentID),
		Text:          c.Text,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}
