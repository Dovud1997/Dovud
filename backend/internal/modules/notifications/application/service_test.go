package application_test

import (
	"context"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/notifications/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/notifications/domain"
	notifpersist "github.com/Dovud1997/Dovud/backend/internal/modules/notifications/infrastructure/persistence"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNotificationCreateMarkReadAndCount(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&notifpersist.NotificationModel{}, &notifpersist.NotificationDeliveryModel{}); err != nil {
		t.Fatal(err)
	}

	svc := application.NewService(notifpersist.NewNotificationRepo(db), nil)
	tenantID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	n, err := svc.Create(ctx, tenantID, application.CreateInput{
		UserID: userID, Type: "order", Title: "New order", Body: "Order submitted", Channel: domain.ChannelInApp,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if n.Channel != domain.ChannelInApp {
		t.Fatalf("channel=%s", n.Channel)
	}

	count, err := svc.UnreadCount(ctx, tenantID, userID)
	if err != nil || count != 1 {
		t.Fatalf("unread count=%d err=%v", count, err)
	}

	list, total, err := svc.ListByUser(ctx, tenantID, userID, true, 1, 20)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list unread: total=%d len=%d err=%v", total, len(list), err)
	}

	if err := svc.MarkRead(ctx, tenantID, userID, n.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	count, err = svc.UnreadCount(ctx, tenantID, userID)
	if err != nil || count != 0 {
		t.Fatalf("after read count=%d err=%v", count, err)
	}

	testN, err := svc.CreateTest(ctx, tenantID, userID)
	if err != nil || testN.Type != "test" {
		t.Fatalf("test notification: err=%v n=%v", err, testN)
	}
	updated, err := svc.MarkAllRead(ctx, tenantID, userID)
	if err != nil || updated != 1 {
		t.Fatalf("mark all: updated=%d err=%v", updated, err)
	}
}

func TestDeviceDeliveryUpsertAndList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:notify-devices-"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&notifpersist.NotificationModel{}, &notifpersist.NotificationDeliveryModel{}); err != nil {
		t.Fatal(err)
	}
	repo := notifpersist.NewNotificationRepo(db)
	svc := application.NewService(repo, nil)
	tenantID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	n, err := svc.Create(ctx, tenantID, application.CreateInput{
		UserID: userID, Type: "promo", Title: "Push me", Body: "Hello", Channel: domain.ChannelPush,
	})
	if err != nil {
		t.Fatal(err)
	}
	dev1, plat1, suf1 := "device-a", "android", "token001"
	dev2, plat2, suf2 := "device-b", "ios", "token002"
	if err := repo.UpsertDeviceDelivery(ctx, &domain.NotificationDelivery{
		NotificationID: n.ID, Channel: domain.ChannelPush, Status: domain.DeliverySent,
		DeviceID: &dev1, Platform: &plat1, TokenSuffix: &suf1,
	}); err != nil {
		t.Fatal(err)
	}
	errMsg := "fcm 404"
	if err := repo.UpsertDeviceDelivery(ctx, &domain.NotificationDelivery{
		NotificationID: n.ID, Channel: domain.ChannelPush, Status: domain.DeliveryFailed,
		Error: &errMsg, DeviceID: &dev2, Platform: &plat2, TokenSuffix: &suf2,
	}); err != nil {
		t.Fatal(err)
	}
	// upsert same device again
	if err := repo.UpsertDeviceDelivery(ctx, &domain.NotificationDelivery{
		NotificationID: n.ID, Channel: domain.ChannelPush, Status: domain.DeliverySent,
		DeviceID: &dev2, Platform: &plat2, TokenSuffix: &suf2,
	}); err != nil {
		t.Fatal(err)
	}

	rows, err := svc.ListDeliveries(ctx, tenantID, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	var deviceRows int
	var sawB bool
	for _, d := range rows {
		if d.DeviceID != nil && *d.DeviceID != "" {
			deviceRows++
			if *d.DeviceID == "device-b" {
				sawB = true
				if d.Status != domain.DeliverySent {
					t.Fatalf("device-b status=%s", d.Status)
				}
			}
		}
	}
	if deviceRows != 2 {
		t.Fatalf("expected 2 device deliveries, got %d (total rows=%d)", deviceRows, len(rows))
	}
	if !sawB {
		t.Fatal("missing device-b delivery")
	}
}
