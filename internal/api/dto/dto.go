package dto

type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	Active    int64   `json:"active"`
	Verified  int64   `json:"verified"`
	Programid string  `json:"programid"`
	CreatedAt string  `json:"created_at"`
	Theme     string  `json:"theme"`
	Private   int64   `json:"private"`
	AvatarUrl *string `json:"avatar_url"`
}

type PublicUserResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	Active    int64   `json:"active"`
	Verified  int64   `json:"verified"`
	Programid string  `json:"programid"`
	CreatedAt string  `json:"created_at"`
	Theme     string  `json:"theme"`
	Private   int64   `json:"private"`
	AvatarUrl *string `json:"avatar_url"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@studmail.w-hs.de"`
	Password string `json:"password" validate:"required,min=8,max=72" example:"secret123"`
	Remember bool   `json:"remember"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email" example:"user@studmail.w-hs.de"`
	Name      string `json:"name" validate:"required,min=2,max=64" example:"Max Mustermann"`
	Password  string `json:"password" validate:"required,min=8,max=72" example:"secret123"`
	Programid string `json:"programid" validate:"required"`
}

type UpdateProfileRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=64" example:"Max Mustermann"`
	Programid string `json:"programid" validate:"required"`
	Theme     string `json:"theme" validate:"required,oneof=light dark system" example:"dark"`
	Private   bool   `json:"private"`
}

// Forum DTOs
type PostResponse struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	AuthorID        string   `json:"author_id"`
	AuthorName      string   `json:"author_name"`
	AuthorAvatarUrl string   `json:"author_avatar_url"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	Pinned          int64    `json:"pinned"`
	Type            string   `json:"type"`
	Programs        []string `json:"programs"`
	Tags            []string `json:"tags"`
	EventDate       *string  `json:"event_date"`
	Location        *string  `json:"location"`
	ImageUrl        *string  `json:"image_url"`
	Links           []string `json:"links"`
	CommentCount    int64    `json:"comment_count"`
	Votes           int64    `json:"votes"`
	UserVote        int64    `json:"user_vote"`
	Active          int64    `json:"active"`
}

type CreatePostRequest struct {
	Title     *string  `json:"title,omitempty" validate:"required,min=3,max=255"`
	Body      *string  `json:"body,omitempty" validate:"required,min=10,max=10000"`
	Type      *string  `json:"type,omitempty" validate:"omitempty,oneof=forum news event"`
	Pinned    *int64   `json:"pinned,omitempty"`
	Programs  []string `json:"programs,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	EventDate *string  `json:"event_date,omitempty"`
	Location  *string  `json:"location,omitempty"`
	ImageURL  *string  `json:"image_url,omitempty"`
	Links     []string `json:"links,omitempty"`
}

type CommentResponse struct {
	ID              string  `json:"id"`
	PostID          string  `json:"post_id"`
	AuthorID        string  `json:"author_id"`
	AuthorName      string  `json:"author_name"`
	AuthorAvatarUrl string  `json:"author_avatar_url"`
	ParentID        *string `json:"parent_id"`
	Text            string  `json:"text"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	Votes           int64   `json:"votes"`
	UserVote        int64   `json:"user_vote"`
}

type CommentRequest struct {
	Text     string  `json:"text" validate:"required,min=1,max=2000"`
	ParentID *string `json:"parent_id,omitempty"`
}

type VoteRequest struct {
	Vote int `json:"vote" validate:"oneof=1 -1 0"`
}

type ExamResponse struct {
	ID           string `json:"id"`
	ProgramID    string `json:"programid"`
	Version      string `json:"version"`
	ModuleID     string `json:"moduleid"`
	ModuleName   string `json:"module_name"`
	ExamDate     string `json:"exam_date"`
	UploadedAt   string `json:"uploaded_at"`
	UploaderName string `json:"uploader_name"`
	Comment      string `json:"comment,omitempty"`
	EditVersion  int64  `json:"edit_version"`
	GroupID      string `json:"group_id"`
	IsLatest     int64  `json:"is_latest"`
}

type ExamAssignment struct {
	ProgramID string `json:"programid" validate:"required"`
	Version   string `json:"version" validate:"required,min=2,max=32"`
	ModuleID  string `json:"moduleid" validate:"required"`
}

type UpdateExamRequest struct {
	ProgramID string `json:"programid" validate:"required"`
	Version   string `json:"version" validate:"required,min=2,max=32"`
	ModuleID  string `json:"moduleid" validate:"required"`
	Date      string `json:"date" validate:"required,datetime=2006-01-02"`
	Comment   string `json:"comment" validate:"max=1000"`
}

type CreateEventRequest struct {
	Title string `validate:"required,min=2,max=255"`
}

type CreateMediaRequest struct {
	EventID     string `validate:"required"`
	Title       string `validate:"max=255"`
	Description string `validate:"max=2000"`
}

type EventResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CoverPath string `json:"cover_path,omitempty"`
	CreatedAt string `json:"created_at"`
}

type MediaResponse struct {
	ID           string `json:"id"`
	EventID      string `json:"event_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	UploadedAt   string `json:"uploaded_at"`
	UploaderName string `json:"uploader_name"`
	MimeType     string `json:"mime_type"`
}

type ActivityResponse struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	UserName   string  `json:"user_name"`
	Type       string  `json:"type"`
	TargetID   string  `json:"target_id"`
	TargetName *string `json:"target_name"`
	CreatedAt  string  `json:"created_at"`
}

type NotificationResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Link      string `json:"link"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

// Search DTOs
type SearchResult struct {
	Type     string `json:"type"` // "exam" or "module" or "user" or "post"
	ID       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	URL      string `json:"url"`
}

type ProgramResponse struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

type ModuleResponse struct {
	ID        string `json:"id"`
	ProgramID string `json:"programid"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type CsrfResponse struct {
	Csrf string `json:"csrf"`
}
