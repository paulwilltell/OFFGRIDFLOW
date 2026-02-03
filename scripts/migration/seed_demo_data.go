package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
)

type seedConfig struct {
	DSN            string
	TenantName     string
	TenantSlug     string
	TenantPlan     string
	UserEmail      string
	UserName       string
	UserPassword   string
	SeedActivities bool
}

func main() {
	cfg := parseFlags()

	if cfg.DSN == "" {
		log.Fatal("missing database DSN (set -dsn or OFFGRIDFLOW_DB_DSN)")
	}

	ctx := context.Background()
	database, err := db.ConnectWithDSN(ctx, cfg.DSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	if err := database.RunMigrations(ctx); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	authStore := auth.NewPostgresStore(database.DB)
	activityStore := ingestion.NewPostgresActivityStore(database.DB)

	tenant, tenantCreated, err := ensureTenant(ctx, authStore, cfg)
	if err != nil {
		log.Fatalf("ensure tenant: %v", err)
	}
	if tenantCreated {
		log.Printf("created tenant: %s (%s)", tenant.Name, tenant.ID)
	} else {
		log.Printf("using existing tenant: %s (%s)", tenant.Name, tenant.ID)
	}

	userCreated, err := ensureUser(ctx, authStore, tenant.ID, cfg)
	if err != nil {
		log.Fatalf("ensure user: %v", err)
	}
	if userCreated {
		log.Printf("created user: %s", cfg.UserEmail)
	} else {
		log.Printf("using existing user: %s", cfg.UserEmail)
	}

	if cfg.SeedActivities {
		seeded, err := ensureActivities(ctx, activityStore, tenant.ID)
		if err != nil {
			log.Fatalf("seed activities: %v", err)
		}
		if seeded {
			log.Printf("seeded demo activities for tenant %s", tenant.ID)
		} else {
			log.Printf("activities already present for tenant %s; skipping seed", tenant.ID)
		}
	}

	log.Println("demo data seed complete")
}

func parseFlags() seedConfig {
	var cfg seedConfig
	flag.StringVar(&cfg.DSN, "dsn", getenv("OFFGRIDFLOW_DB_DSN", getenv("DATABASE_URL", "")), "Postgres DSN")
	flag.StringVar(&cfg.TenantName, "tenant-name", getenv("OFFGRIDFLOW_DEMO_TENANT_NAME", "OffGridFlow Demo"), "Tenant name")
	flag.StringVar(&cfg.TenantSlug, "tenant-slug", getenv("OFFGRIDFLOW_DEMO_TENANT_SLUG", ""), "Tenant slug (optional)")
	flag.StringVar(&cfg.TenantPlan, "tenant-plan", getenv("OFFGRIDFLOW_DEMO_TENANT_PLAN", "enterprise"), "Tenant plan (free|pro|enterprise)")
	flag.StringVar(&cfg.UserEmail, "user-email", getenv("OFFGRIDFLOW_DEMO_USER_EMAIL", "demo@offgridflow.com"), "Demo user email")
	flag.StringVar(&cfg.UserName, "user-name", getenv("OFFGRIDFLOW_DEMO_USER_NAME", "Demo Admin"), "Demo user name")
	flag.StringVar(&cfg.UserPassword, "user-password", getenv("OFFGRIDFLOW_DEMO_USER_PASSWORD", ""), "Demo user password (required if user does not exist)")
	flag.BoolVar(&cfg.SeedActivities, "seed-activities", true, "Seed demo activities")
	flag.Parse()

	return cfg
}

func ensureTenant(ctx context.Context, store auth.Store, cfg seedConfig) (*auth.Tenant, bool, error) {
	if cfg.TenantName == "" {
		return nil, false, errors.New("tenant name is required")
	}

	if tenant, err := store.GetTenantByName(ctx, cfg.TenantName); err == nil {
		return tenant, false, nil
	} else if !errors.Is(err, auth.ErrTenantNotFound) {
		return nil, false, err
	}

	now := time.Now()
	tenant := &auth.Tenant{
		ID:        uuid.NewString(),
		Name:      cfg.TenantName,
		Slug:      buildSlug(cfg.TenantSlug, cfg.TenantName),
		Plan:      cfg.TenantPlan,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.CreateTenant(ctx, tenant); err != nil {
		return nil, false, err
	}

	return tenant, true, nil
}

func ensureUser(ctx context.Context, store auth.Store, tenantID string, cfg seedConfig) (bool, error) {
	if cfg.UserEmail == "" {
		return false, errors.New("user email is required")
	}
	if tenantID == "" {
		return false, errors.New("tenant ID is required")
	}

	if _, err := store.GetUserByEmail(ctx, cfg.UserEmail); err == nil {
		return false, nil
	} else if !errors.Is(err, auth.ErrUserNotFound) {
		return false, err
	}

	if cfg.UserPassword == "" {
		return false, errors.New("user password is required to create demo user")
	}

	hash, err := auth.HashPassword(cfg.UserPassword)
	if err != nil {
		return false, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &auth.User{
		ID:            uuid.NewString(),
		Email:         strings.ToLower(strings.TrimSpace(cfg.UserEmail)),
		Name:          cfg.UserName,
		TenantID:      tenantID,
		Role:          "admin",
		Roles:         []string{"admin"},
		PasswordHash:  hash,
		IsActive:      true,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := store.CreateUser(ctx, user); err != nil {
		return false, err
	}

	return true, nil
}

func ensureActivities(ctx context.Context, store ingestion.ActivityStore, tenantID string) (bool, error) {
	if tenantID == "" {
		return false, errors.New("tenant ID is required")
	}

	existing, err := store.ListByOrg(ctx, tenantID)
	if err != nil {
		return false, err
	}
	if len(existing) > 0 {
		return false, nil
	}

	demoStore := ingestion.NewInMemoryActivityStore()
	demoStore.SeedDemoData()
	activities, err := demoStore.List(ctx)
	if err != nil {
		return false, err
	}

	for i := range activities {
		activities[i].ID = ""
		activities[i].OrgID = tenantID
	}

	if err := store.SaveBatch(ctx, activities); err != nil {
		return false, err
	}

	return true, nil
}

func buildSlug(explicit, name string) string {
	slug := strings.TrimSpace(explicit)
	if slug != "" {
		return slug
	}

	if name == "" {
		name = "org"
	}

	var b strings.Builder
	lastDash := false
	for _, c := range strings.ToLower(name) {
		switch {
		case c >= 'a' && c <= 'z':
			b.WriteRune(c)
			lastDash = false
		case c >= '0' && c <= '9':
			b.WriteRune(c)
			lastDash = false
		case c == ' ' || c == '-' || c == '_':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	slug = strings.Trim(b.String(), "-")
	if slug == "" {
		slug = "org"
	}

	return slug + "-" + uuid.New().String()[:8]
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
