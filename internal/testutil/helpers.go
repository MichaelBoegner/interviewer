package testutil

import "testing"

func CreateTestUserAndJWT(t *testing.T) (int, string) {
	t.Helper()

	// 1. Register a new user via POST /api/users
	// 2. Login via POST /api/auth/login
	// 3. Extract JWT + userID from response

	return userID, jwtToken
}
