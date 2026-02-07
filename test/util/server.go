package testutil

// // SetupTest runs before each test.
// func (s *TestServices) SetupTest() {
// 	// Clean up test data before each test
// 	ctx := context.Background()
// 	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_device"`)
// 	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_pin"`)
// 	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_profile"`)
// 	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user_file"`)
// 	_, _ = s.DBPool.Exec(ctx, `DELETE FROM "user"`)

// 	// Clean up Redis keys
// 	iter := s.Redis.Scan(ctx, 0, "test:*", 100).Iterator()
// 	for iter.Next(ctx) {
// 		s.Redis.Del(ctx, iter.Val())
// 	}
// }

// // CreateTestUserViaService creates a test user through the service layer.
// // This is the recommended way to create users in tests as it exercises
// // the full service logic including hashing.
// func CreateTestUserViaService(ctx context.Context, svc *TestServices, username, email, password string) (*TestUser, error) {
// 	user, err := svc.UserService.Create(ctx, &params.UserCreateParam{
// 		Username: username,
// 		Email:    email,
// 		Password: password,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create test user: %w", err)
// 	}

// 	return &TestUser{
// 		User:     user,
// 		Password: password,
// 	}, nil
// }

// // TestUser wraps a User with the plain-text password for testing.
// type TestUser struct {
// 	User     *model.User
// 	Password string
// }

// // CleanupTestData cleans up all test data from database and Redis.
// func CleanupTestData(ctx context.Context, svc *TestServices) error {
// 	// Clean up database
// 	if err := SetupTestDatabase(ctx, svc.DBPool); err != nil {
// 		return fmt.Errorf("failed to cleanup database: %w", err)
// 	}

// 	// Clean up Redis
// 	if err := SetupTestRedis(ctx, svc.Redis); err != nil {
// 		return fmt.Errorf("failed to cleanup Redis: %w", err)
// 	}

// 	return nil
// }

// // WaitForInfrastructure waits for both database and Redis to be ready.
// func WaitForInfrastructure(ctx context.Context, dbURL, redisURL string, maxAttempts int) error {
// 	// Wait for database
// 	if err := WaitForDatabase(ctx, dbURL, maxAttempts); err != nil {
// 		return fmt.Errorf("database not ready: %w", err)
// 	}

// 	// Wait for Redis
// 	redisClient := redis.NewClient(&redis.Options{
// 		Addr: redisURL,
// 		DB:   0,
// 	})
// 	defer redisClient.Close()

// 	if err := WaitForRedis(ctx, redisClient, maxAttempts); err != nil {
// 		return fmt.Errorf("redis not ready: %w", err)
// 	}

// 	return nil
// }
