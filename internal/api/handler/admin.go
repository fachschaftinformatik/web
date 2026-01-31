package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
)

// @Summary Create a program
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body dto.ProgramResponse true "Program Data"
// @Success 201 {object} dto.ProgramResponse
// @Router /admin/programs [post]
func (s *Server) PostAdminPrograms(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name     string   `json:"name" validate:"required"`
		Versions []string `json:"versions" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "bad_request", "Invalid JSON", http.StatusBadRequest)
		return
	}

	prog, err := s.DB.CreateProgram(r.Context(), database.CreateProgramParams{
		ID:   int64(id.New()),
		Name: payload.Name,
	})
	if err != nil {
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	for _, v := range payload.Versions {
		_ = s.DB.CreateProgramVersion(r.Context(), database.CreateProgramVersionParams{
			ProgramID: prog.ID,
			Name:      v,
		})
	}

	s.RespondJSON(w, http.StatusCreated, dto.ProgramResponse{
		ID:       id.ID(prog.ID),
		Name:     prog.Name,
		Versions: payload.Versions,
	})
}

// @Summary Create a module
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body dto.ModuleResponse true "Module Data"
// @Success 201 {object} dto.ModuleResponse
// @Router /admin/modules [post]
func (s *Server) PostAdminModules(w http.ResponseWriter, r *http.Request) {
	var payload dto.ModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "bad_request", "Invalid JSON", http.StatusBadRequest)
		return
	}

	mod, err := s.DB.CreateModule(r.Context(), database.CreateModuleParams{
		ID:        int64(id.New()),
		ProgramID: int64(payload.ProgramID),
		Name:      payload.Name,
		Alias:     &payload.Alias,
	})
	if err != nil {
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusCreated, dto.ModuleResponse{
		ID:        id.ID(mod.ID),
		ProgramID: id.ID(mod.ProgramID),
		Name:      mod.Name,
		Alias:     *mod.Alias,
	})
}

// @Summary Get all user roles
// @Tags Admin
// @ID getAdminRefRoles
// @Success 200 {array} string
// @Router /admin/ref/roles [get]
func (s *Server) GetAdminRefRoles(w http.ResponseWriter, r *http.Request) {
	roles := []string{"admin", "editor", "user"}
	s.RespondJSON(w, http.StatusOK, roles)
}

// @Summary Get all discussion types
// @Tags Admin
// @ID getAdminRefDiscussionTypes
// @Success 200 {array} string
// @Router /admin/ref/discussion-types [get]
func (s *Server) GetAdminRefDiscussionTypes(w http.ResponseWriter, r *http.Request) {
	types := []string{"discussion", "news", "event"}
	s.RespondJSON(w, http.StatusOK, types)
}

// @Summary Get admin dashboard stats
// @Tags Admin
// @ID getAdminStats
// @Success 200 {object} dto.AdminStatsResponse
// @Router /admin/stats [get]
func (s *Server) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.DB.GetAdminStats(r.Context())
	if err != nil {
		s.JsonError(w, "database_error", err.Error(), http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.AdminStatsResponse{
		UserCount:     stats.UserCount,
		PostCount:     stats.PostCount,
		ArchiveCount:  stats.ArchiveCount,
		EventCount:    stats.EventCount,
		ModuleCount:   stats.ModuleCount,
		ProgramCount:  stats.ProgramCount,
		ActivityCount: stats.ActivityCount,
		SessionCount:  stats.SessionCount,
	})
}

// @Summary Get admin dashboard data
// @Tags Admin
// @ID getAdminDashboard
// @Success 200 {object} dto.AdminDashboardResponse
// @Router /admin/dashboard [get]
func (s *Server) GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, _ := s.DB.GetAdminStats(ctx)
	growth, _ := s.DB.GetDailyUserGrowth(ctx)
	trend, _ := s.DB.GetDailyActivityTrend(ctx)
	examGrowth, _ := s.DB.GetDailyExamGrowth(ctx)
	discGrowth, _ := s.DB.GetDailyDiscussionGrowth(ctx)
	moduleGrowth, _ := s.DB.GetDailyModuleGrowth(ctx)
	programGrowth, _ := s.DB.GetDailyProgramGrowth(ctx)
	sessionTrend, _ := s.DB.GetDailySessionTrend(ctx)
	dist, _ := s.DB.GetUserProgramDistribution(ctx)
	activities, _ := s.DB.ListAllActivities(ctx, database.ListAllActivitiesParams{
		Limit:  10,
		Offset: 0,
	})

	growthDto := make([]dto.DashboardTrendItem, len(growth))
	for i, g := range growth {
		dateStr := ""
		if s, ok := g.Date.(string); ok {
			dateStr = s
		} else if g.Date != nil {
			dateStr = fmt.Sprint(g.Date)
		}
		growthDto[i] = dto.DashboardTrendItem{
			Date:  dateStr,
			Count: g.Count,
		}
	}

	trendDto := make([]dto.DashboardTrendItem, len(trend))
	for i, t := range trend {
		dateStr := ""
		if s, ok := t.Date.(string); ok {
			dateStr = s
		} else if t.Date != nil {
			dateStr = fmt.Sprint(t.Date)
		}
		trendDto[i] = dto.DashboardTrendItem{
			Date:  dateStr,
			Count: t.Count,
		}
	}

	examDto := make([]dto.DashboardTrendItem, len(examGrowth))
	for i, g := range examGrowth {
		dateStr := ""
		if s, ok := g.Date.(string); ok {
			dateStr = s
		} else if g.Date != nil {
			dateStr = fmt.Sprint(g.Date)
		}
		examDto[i] = dto.DashboardTrendItem{
			Date:  dateStr,
			Count: g.Count,
		}
	}

	discDto := make([]dto.DashboardTrendItem, len(discGrowth))
	for i, g := range discGrowth {
		dateStr := ""
		if s, ok := g.Date.(string); ok {
			dateStr = s
		} else if g.Date != nil {
			dateStr = fmt.Sprint(g.Date)
		}
		discDto[i] = dto.DashboardTrendItem{
			Date:  dateStr,
			Count: g.Count,
		}
	}

	moduleGrowthDto := make([]dto.DashboardTrendItem, len(moduleGrowth))
	for i, g := range moduleGrowth {
		dateStr := ""
		if s, ok := g.Date.(string); ok {
			dateStr = s
		} else if g.Date != nil {
			dateStr = fmt.Sprint(g.Date)
		}
		moduleGrowthDto[i] = dto.DashboardTrendItem{
			Date:  dateStr,
			Count: g.Count,
		}
	}

	programGrowthDto := make([]dto.DashboardTrendItem, len(programGrowth))
	for i, g := range programGrowth {
		dateStr := ""
		if s, ok := g.Date.(string); ok {
			dateStr = s
		} else if g.Date != nil {
			dateStr = fmt.Sprint(g.Date)
		}
		programGrowthDto[i] = dto.DashboardTrendItem{
			Date:  dateStr,
			Count: g.Count,
		}
	}

	sessionDto := make([]dto.DashboardTrendItem, len(sessionTrend))
	for i, s := range sessionTrend {
		dateStr := ""
		if str, ok := s.Date.(string); ok {
			dateStr = str
		} else if s.Date != nil {
			dateStr = fmt.Sprint(s.Date)
		}
		sessionDto[i] = dto.DashboardTrendItem{
			Date:  dateStr,
			Count: s.Count,
		}
	}

	distDto := make([]dto.ProgramDistributionItem, len(dist))
	for i, d := range dist {
		distDto[i] = dto.ProgramDistributionItem{
			Name:  d.ProgramName,
			Value: d.UserCount,
		}
	}

	activitiesDto := make([]dto.ActivityResponse, len(activities))
	for i, a := range activities {
		activitiesDto[i] = dto.ActivityResponse{
			ID:         id.ID(a.ID),
			UserID:     id.ID(a.UserID),
			UserName:   a.UserName,
			Type:       a.Type,
			TargetID:   a.TargetID,
			TargetName: a.TargetName,
			CreatedAt:  a.CreatedAt,
		}
	}

	s.RespondJSON(w, http.StatusOK, dto.AdminDashboardResponse{
		Stats: dto.AdminStatsResponse{
			UserCount:     stats.UserCount,
			PostCount:     stats.PostCount,
			ArchiveCount:  stats.ArchiveCount,
			EventCount:    stats.EventCount,
			ModuleCount:   stats.ModuleCount,
			ProgramCount:  stats.ProgramCount,
			ActivityCount: stats.ActivityCount,
			SessionCount:  stats.SessionCount,
		},
		UserGrowthTrend:       growthDto,
		ExamGrowthTrend:       examDto,
		DiscussionGrowthTrend: discDto,
		ModuleGrowthTrend:     moduleGrowthDto,
		ProgramGrowthTrend:    programGrowthDto,
		ActivityTrend:         trendDto,
		SessionTrend:          sessionDto,
		ProgramDistribution:   distDto,
		RecentActivities:      activitiesDto,
	})
}
