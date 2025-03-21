// package main_test

// import (
// 	"log"
// 	"os"
// 	"testing"

// 	"github.com/joho/godotenv"
// 	"github.com/michaelboegner/interviewer/internal/testutil"
// )

// func TestMain(m *testing.M) {
// 	log.Println("Loading environment variables...")
// 	err := godotenv.Load(".env.test")
// 	if err != nil {
// 		log.Fatalf("Error loading .env.test file: %v", err)
// 	}

// 	log.Println("Initializing test server...")
// 	testutil.InitTestServer()

// 	// 🚨 Check `TestServerURL` before running any tests
// 	if testutil.TestServerURL == "" {
// 		log.Fatal("TestMain: TestServerURL is empty! The server did not start properly.")
// 	}

// 	log.Printf("TestMain: Test server started successfully at: %s", testutil.TestServerURL)

// 	code := m.Run()

// 	log.Println("Stopping test server...")
// 	testutil.StopTestServer()

// 	os.Exit(code)
// }
