package dto

import (
	"github.com/fachschaftinformatik/web/internal/id"
)

type UserResponse struct {
	ID        id.ID   `json:"id" swaggertype:"string"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	Active    int64   `json:"active"`
	Verified  int64   `json:"verified"`
	ProgramID *id.ID  `json:"program_id" swaggertype:"string"`
	CreatedAt string  `json:"created_at"`
	Theme     string  `json:"theme"`
	Private   int64   `json:"private"`
	AvatarUrl *string `json:"avatar_url"`
}

type PublicUserResponse struct {
	ID        id.ID   `json:"id" swaggertype:"string"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	Active    int64   `json:"active"`
	Verified  int64   `json:"verified"`
	ProgramID *id.ID  `json:"program_id" swaggertype:"string"`
	CreatedAt string  `json:"created_at"`
	Theme     string  `json:"theme"`
	Private   int64   `json:"private"`
	AvatarUrl *string `json:"avatar_url"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@fsv-wh.de"`
	Password string `json:"password" validate:"required,min=8,max=72" example:"secret123"`
	Remember bool   `json:"remember"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email" example:"user@fsv-wh.de"`
	Name      string `json:"name" validate:"required,min=2,max=64" example:"Max Mustermann"`
	Password  string `json:"password" validate:"required,min=8,max=72" example:"secret123"`
	ProgramID *id.ID `json:"program_id" validate:"omitempty" swaggertype:"string"`
}

type UpdateProfileRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=64" example:"Max Mustermann"`
	ProgramID *id.ID `json:"program_id" validate:"omitempty" swaggertype:"string"`
	Theme     string `json:"theme" validate:"required,oneof=light dark system" example:"dark"`
	Private   bool   `json:"private"`
}

// Discussion DTOs
type PostProgramResponse struct {
	ID   id.ID  `json:"id" swaggertype:"string"`
	Name string `json:"name"`
}

type DiscussionPostResponse struct {
	ID            id.ID                 `json:"id" swaggertype:"string"`
	Title         string                `json:"title"`
	Body          string                `json:"body"`
	UserID        id.ID                 `json:"user_id" swaggertype:"string"`
	UserName      string                `json:"user_name"`
	UserAvatarUrl string                `json:"user_avatar_url"`
	CreatedAt     string                `json:"created_at"`
	UpdatedAt     string                `json:"updated_at"`
	Pinned        int64                 `json:"pinned"`
	Type          string                `json:"type"`
	Programs      []PostProgramResponse `json:"programs"`
	Tags          []string              `json:"tags"`
	EventDate     *string               `json:"event_date"`
	Location      *string               `json:"location"`
	ImageUrl      *string               `json:"image_url"`
	Links         []Link                `json:"links"`
	CommentCount  int64                 `json:"comment_count"`
	Votes         int64                 `json:"votes"`
	UserVote      int64                 `json:"user_vote"`
}

type Link struct {
	URL   string `json:"url"`
	Label string `json:"label,omitempty"`
}

type CreateDiscussionPostRequest struct {
	Title     string   `json:"title" validate:"required,min=3,max=255"`
	Body      string   `json:"body" validate:"required,min=10,max=10000"`
	Type      string   `json:"type" validate:"omitempty,oneof=discussion news event"`
	Pinned    bool     `json:"pinned"`
	Programs  []id.ID  `json:"programs" swaggertype:"array,string"`
	Tags      []string `json:"tags"`
	EventDate *string  `json:"event_date,omitempty"`
	Location  *string  `json:"location,omitempty"`
	ImageURL  *string  `json:"image_url,omitempty"`
	Links     []Link   `json:"links"`
}

type UpdateDiscussionPostRequest struct {
	Title     *string  `json:"title" validate:"omitempty,min=3,max=255"`
	Body      *string  `json:"body" validate:"omitempty,min=10,max=10000"`
	Type      *string  `json:"type" validate:"omitempty,oneof=discussion news event"`
	Pinned    *bool    `json:"pinned"`
	Programs  []id.ID  `json:"programs" swaggertype:"array,string"`
	Tags      []string `json:"tags"`
	EventDate *string  `json:"event_date,omitempty"`
	Location  *string  `json:"location,omitempty"`
	ImageURL  *string  `json:"image_url,omitempty"`
	Links     []Link   `json:"links"`
}

type DiscussionCommentResponse struct {
	ID            id.ID  `json:"id" swaggertype:"string"`
	PostID        id.ID  `json:"post_id" swaggertype:"string"`
	UserID        id.ID  `json:"user_id" swaggertype:"string"`
	UserName      string `json:"user_name"`
	UserAvatarUrl string `json:"user_avatar_url"`
	ParentID      *id.ID `json:"parent_id" swaggertype:"string"`
	Text          string `json:"text"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Votes         int64  `json:"votes"`
	UserVote      int64  `json:"user_vote"`
}

type DiscussionCommentRequest struct {
	Text     string `json:"text" validate:"required,min=1,max=2000"`
	ParentID *id.ID `json:"parent_id,omitempty" swaggertype:"string"`
}

type VoteRequest struct {
	Vote int `json:"vote" validate:"oneof=1 -1 0"`
}

// Archive DTOs
type ArchiveEntryResponse struct {
	ID           id.ID  `json:"id" swaggertype:"string"`
	ModuleID     id.ID  `json:"module_id" swaggertype:"string"`
	ModuleName   string `json:"module_name"`
	ProgramID    id.ID  `json:"program_id" swaggertype:"string"`
	Version      string `json:"version"`
	ExamDate     string `json:"exam_date"`
	CreatedAt    string `json:"created_at"`
	UploaderName string `json:"uploader_name"`
	Comment      string `json:"comment,omitempty"`
	EditVersion  int64  `json:"edit_version"`
	FileID       id.ID  `json:"file_id" swaggertype:"string"`
}

type CreateArchiveEntryRequest struct {
	ModuleID id.ID  `json:"module_id" validate:"required" swaggertype:"string"`
	Version  string `json:"version" validate:"required,min=2,max=32"`
	Date     string `json:"date" validate:"required,datetime=2006-01-02"`
	Comment  string `json:"comment" validate:"max=1000"`
}

type UpdateArchiveEntryRequest struct {
	ModuleID id.ID  `json:"module_id" validate:"required" swaggertype:"string"`
	Version  string `json:"version" validate:"required,min=2,max=32"`
	Date     string `json:"date" validate:"required,datetime=2006-01-02"`
	Comment  string `json:"comment" validate:"max=1000"`
}

// Event DTOs
type EventResponse struct {
	ID        id.ID   `json:"id" swaggertype:"string"`
	Title     string  `json:"title"`
	CoverPath *string `json:"cover_path,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type CreateEventRequest struct {
	Title string `json:"title" validate:"required,min=2,max=255"`
}

type MediaResponse struct {
	ID           id.ID  `json:"id" swaggertype:"string"`
	EventID      id.ID  `json:"event_id" swaggertype:"string"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
	UploaderName string `json:"uploader_name"`
	MimeType     string `json:"mime_type"`
}

type CreateMediaRequest struct {
	EventID     id.ID  `json:"event_id" validate:"required" swaggertype:"string"`
	Title       string `json:"title" validate:"max=255"`
	Description string `json:"description" validate:"max=2000"`
}

type ActivityResponse struct {
	ID         id.ID   `json:"id" swaggertype:"string"`
	UserID     id.ID   `json:"user_id" swaggertype:"string"`
	UserName   string  `json:"user_name"`
	Type       string  `json:"type"`
	TargetID   string  `json:"target_id"`
	TargetName *string `json:"target_name"`
	CreatedAt  string  `json:"created_at"`
}

type NotificationResponse struct {
	ID        id.ID  `json:"id" swaggertype:"string"`
	UserID    id.ID  `json:"user_id" swaggertype:"string"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Link      string `json:"link"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

// Search DTOs
type SearchResult struct {
	Type     string `json:"type"` // "archive" or "module" or "user" or "discussion"
	ID       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	URL      string `json:"url"`
}

type ProgramResponse struct {
	ID       id.ID    `json:"id" swaggertype:"string"`
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

type ModuleResponse struct {
	ID        id.ID  `json:"id" swaggertype:"string"`
	ProgramID id.ID  `json:"program_id" swaggertype:"string"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type CsrfResponse struct {
	Csrf           string `json:"csrf"`
	SignupsEnabled bool   `json:"signups_enabled"`
}

type ListResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
}

type OfficeStatusResponse struct {
	Occupied bool `json:"occupied"`
}

type UpdateOfficeStatusRequest struct {
	Occupied bool `json:"occupied"`
}
