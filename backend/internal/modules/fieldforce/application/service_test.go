package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/application"
	ffpersist "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type memRecorder struct {
	n    int
	last string
}

func (m *memRecorder) RecordChange(_ context.Context, _ uuid.UUID, entityType, _ string, _ int64, _ bool, _ any) error {
	m.n++
	m.last = entityType
	return nil
}

func TestCheckInCheckOutAndGPS(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&ffpersist.SalesAgentModel{},
		&ffpersist.RouteModel{},
		&ffpersist.RouteStopModel{},
		&ffpersist.VisitModel{},
		&ffpersist.VisitPhotoModel{},
		&ffpersist.VisitCommentModel{},
		&ffpersist.GpsTrackModel{},
	); err != nil {
		t.Fatal(err)
	}

	rec := &memRecorder{}
	svc := application.NewService(
		ffpersist.NewAgentRepo(db),
		ffpersist.NewRouteRepo(db),
		ffpersist.NewVisitRepo(db),
		ffpersist.NewGpsRepo(db),
	).WithSync(rec)
	tenantID := uuid.New()
	userID := uuid.New()
	branchID := uuid.New()
	customerID := uuid.New()
	ctx := context.Background()

	agent, err := svc.CreateAgent(ctx, tenantID, application.AgentInput{
		UserID: userID, BranchID: branchID, EmployeeCode: "AG-9",
	})
	if err != nil {
		t.Fatalf("agent: %v", err)
	}

	route, err := svc.CreateRoute(ctx, tenantID, application.RouteInput{
		AgentID: agent.ID, Date: time.Now().UTC().Format("2006-01-02"), Name: "R1",
	})
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	route, err = svc.SetStops(ctx, tenantID, route.ID, []application.RouteStopInput{
		{CustomerID: customerID, Sequence: 1},
	})
	if err != nil || len(route.Stops) != 1 {
		t.Fatalf("stops: err=%v stops=%v", err, route.Stops)
	}

	lat, lng := 41.3, 69.2
	visit, err := svc.CheckIn(ctx, tenantID, application.CheckInInput{
		AgentID: agent.ID, CustomerID: customerID, RouteStopID: &route.Stops[0].ID, Lat: &lat, Lng: &lng,
	})
	if err != nil {
		t.Fatalf("checkin: %v", err)
	}

	out, err := svc.CheckOut(ctx, tenantID, visit.ID, application.CheckOutInput{
		Lat: &lat, Lng: &lng, Result: "success",
	})
	if err != nil || out.EndedAt == nil {
		t.Fatalf("checkout: err=%v visit=%+v", err, out)
	}

	points, err := svc.UploadPoints(ctx, tenantID, []application.GpsPointInput{{
		AgentID: agent.ID, VisitID: &visit.ID, Lat: lat, Lng: lng, RecordedAt: time.Now().UTC(),
	}})
	if err != nil || len(points) == 0 {
		t.Fatalf("gps: err=%v points=%v", err, points)
	}
	if rec.last != "gps_point" {
		t.Fatalf("expected gps_point sync fan-out, last=%s n=%d", rec.last, rec.n)
	}
	live, err := svc.LivePosition(ctx, tenantID, agent.ID)
	if err != nil || live == nil {
		t.Fatalf("live: err=%v", err)
	}
}
