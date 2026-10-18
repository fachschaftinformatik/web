package site

import (
	"slices"

	"github.com/fachschaftinformatik/web/model/events"
	"github.com/fachschaftinformatik/web/model/exams"
	"github.com/fachschaftinformatik/web/model/media"
	"github.com/fachschaftinformatik/web/model/meetings"
	"github.com/fachschaftinformatik/web/model/members"
	"github.com/fachschaftinformatik/web/model/organization"
	"github.com/fachschaftinformatik/web/model/posts"
	"github.com/fachschaftinformatik/web/model/users"
	"github.com/fachschaftinformatik/web/view"
)

// Mapping belongs to the controller: views receive data, never model behaviour.
func postView(p posts.Post) view.Post {
	return view.Post{ID: p.ID, AuthorPublicID: p.AuthorPublicID, AuthorID: p.AuthorID,
		AuthorPublic: p.AuthorPublic, Title: p.Title, Body: p.Body, Author: p.Author,
		Created: p.Created, Type: p.Type, BodyFormat: p.BodyFormat, Comments: p.Comments,
		Votes: p.Votes, UserVote: p.UserVote, Tags: slices.Clone(p.Tags)}
}
func commentView(c posts.Comment) view.Comment {
	return view.Comment{ID: c.ID, AuthorPublicID: c.AuthorPublicID, AuthorPublic: c.AuthorPublic,
		Votes: c.Votes, UserVote: c.UserVote, AuthorID: c.AuthorID, CanEdit: c.CanEdit,
		Author: c.Author, Text: c.Text, Created: c.Created, Updated: c.Updated}
}
func threadView(t posts.Thread) view.Thread {
	return view.Thread{Post: postView(t.Post), Updated: t.Updated, EventDate: t.EventDate,
		Location: t.Location, Replies: mapRecords(t.Replies, commentView)}
}
func eventView(e events.Event) view.Event {
	return view.Event{ID: e.ID, Title: e.Title, Body: e.Body, BodyFormat: e.BodyFormat,
		EventDate: e.EventDate, Location: e.Location, Created: e.Created, Updated: e.Updated,
		LegacyPostPublicID: e.LegacyPostPublicID, AlbumPublicID: e.AlbumPublicID,
		AlbumCoverPublicID: e.AlbumCoverPublicID, CoverKey: e.CoverKey, CoverAlt: e.CoverAlt,
		CoverWidth: e.CoverWidth, CoverHeight: e.CoverHeight, SourceURL: e.SourceURL,
		SourcePublishedAt: e.SourcePublishedAt, SourceStatus: e.SourceStatus}
}
func assignmentView(a exams.Assignment) view.Assignment {
	return view.Assignment{ModuleID: a.ModuleID, ProgramID: a.ProgramID, Version: a.Version,
		Title: a.Title, Program: a.Program, Degree: exams.DegreeLabel(a.Degree), Value: a.Key()}
}
func examsFilterView(f exams.Filter) view.ExamsFilter {
	return view.ExamsFilter{Query: f.Query, Programs: slices.Clone(f.Programs),
		Versions: slices.Clone(f.Versions), Degrees: slices.Clone(f.Degrees)}
}
func examsFilterModel(f view.ExamsFilter) exams.Filter {
	return exams.Filter{Query: f.Query, Programs: slices.Clone(f.Programs),
		Versions: slices.Clone(f.Versions), Degrees: slices.Clone(f.Degrees)}
}
func examView(e exams.Exam) view.Exam {
	return view.Exam{ID: e.ID, Title: e.Title, Program: e.Program, Version: e.Version,
		ExamDate: e.Date, Description: e.Description, Degree: e.Degree, ModuleID: e.ModuleID,
		ProgramID: e.ProgramID, Pages: e.Pages, Number: e.Number, Bytes: e.Bytes,
		Assignments: mapRecords(e.Assignments, assignmentView)}
}
func moduleView(m exams.Module) view.Module {
	return view.Module{ID: m.ID, Title: m.Title, ProgramID: m.ProgramID, Program: m.Program,
		StudyProgram: m.StudyProgram, Degree: m.Degree, Exams: mapRecords(m.Exams, examView)}
}
func submissionView(s exams.Submission) view.Submission {
	return view.Submission{ID: s.ID, Name: s.Name, Title: s.Title, Program: s.Program,
		ModuleID: s.ModuleID, Version: s.Version, Date: s.Date, Comment: s.Comment,
		RevisionOf: s.RevisionOf, UserPublicID: s.UserPublicID, UserID: s.UserID,
		Status: view.SubmissionStatus(s.Status), CreatedAt: s.CreatedAt,
		SubmittedAt: clonePointer(s.SubmittedAt), ReviewedAt: clonePointer(s.ReviewedAt),
		Documents: mapRecords(s.Documents, examView), Count: s.Count, ModuleCount: s.ModuleCount,
		Assignments: mapRecords(s.Assignments, assignmentView), SharedAssignments: s.SharedAssignments,
		EditToken: s.EditToken}
}
func imageView(i media.Image) view.Image {
	return view.Image{ID: i.ID, Title: i.Title, Description: i.Description,
		Width: i.Width, Height: i.Height, AlbumCover: i.AlbumCover}
}
func albumView(a media.Album) view.Album {
	return view.Album{ID: a.ID, Title: a.Title, OriginalCount: a.OriginalCount,
		Images: mapRecords(a.Images, imageView), Cover: imageView(a.Cover), PageSize: media.PageSize}
}
func documentView(d meetings.Document) view.Document {
	return view.Document{Kind: d.Kind, File: d.File, Note: d.Note, SHA256: d.SHA256,
		PublicID: d.PublicID, Bytes: d.Bytes, Sources: slices.Clone(d.Sources), Stored: d.Stored}
}
func meetingView(m meetings.Meeting) view.Meeting {
	return view.Meeting{Year: m.Year, Number: m.Number, Date: m.Date, Documents: mapRecords(m.Documents, documentView)}
}
func publicationView(p meetings.Publication) view.Publication {
	return view.Publication{ID: p.ID, Owner: p.Owner, Date: p.Date, Kind: p.Kind,
		Filename: p.Filename, PublicID: p.PublicID, Number: p.Number, Bytes: len(p.Data), Published: p.Published}
}
func memberView(m members.Member) view.Member {
	return view.Member{ID: m.ID, Name: m.Name, Email: m.Email, PublicID: m.PublicID,
		UserID: m.UserID, UserPublicID: m.UserPublicID, AccountEmail: m.AccountEmail, Confirmed: m.Confirmed}
}
func memberRoleView(r members.MemberRole) view.MemberRole {
	return view.MemberRole{ID: r.ID, Group: r.Group, Title: r.Title, Email: r.Email, PublicID: r.PublicID}
}
func teamGroupView(g members.Group) view.TeamGroup {
	return view.TeamGroup{ID: g.ID, Title: g.Title, Sections: mapRecords(g.Sections, func(s members.TeamSection) view.TeamSection {
		return view.TeamSection{Title: s.Title, Group: s.Group, SharedEmail: s.SharedEmail(),
			Members: mapRecords(s.Members, func(m members.TeamMember) view.TeamMember {
				return view.TeamMember{Name: m.Name, Initials: m.Initials, Role: m.Role, Email: m.Email}
			})}
	})}
}
func userView(u users.User) view.User {
	return view.User{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, PublicID: u.PublicID,
		Program: u.Program, Degree: u.Degree, Active: u.Active, Verified: u.Verified}
}
func programView(p users.Program) view.Program { return view.Program{ID: p.ID, Name: p.Name} }
func linkView(l organization.Link) view.Link   { return view.Link{Title: l.Title, URL: l.URL} }
func organizationView(d organization.Details) view.Organization {
	return view.Organization{Name: d.Name, FormalName: d.FormalName, ShortName: d.ShortName,
		Institution: d.Institution, City: d.City, Street: d.Street, PostalCode: d.PostalCode,
		Country: d.Country, Office: d.Office, ContactEmail: d.ContactEmail, ChairRole: d.ChairRole,
		PublicityRole: d.PublicityRole, LegalUpdated: d.LegalUpdated, HostingName: d.HostingName,
		HostingURL: d.HostingURL, PrivacyURL: d.PrivacyURL, SupervisionName: d.SupervisionName,
		SupervisionURL: d.SupervisionURL, PrivacyAuthorityName: d.PrivacyAuthorityName,
		PrivacyAuthorityURL: d.PrivacyAuthorityURL, SocialLinks: mapRecords(d.SocialLinks, linkView),
		ElectionLinks: mapRecords(d.ElectionLinks, linkView), RegulationLinks: mapRecords(d.RegulationLinks, linkView)}
}
func mapRecords[T, U any](values []T, convert func(T) U) []U {
	if values == nil {
		return nil
	}
	result := make([]U, len(values))
	for i, v := range values {
		result[i] = convert(v)
	}
	return result
}
func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
