
# Interviewer App

### A Mock Technical Interview Application built with Go

Welcome to **Interviewerio**, a backend system designed to simulate mock technical interviews. This project is built to help users practice for technical job interviews by interacting with dynamic, pre-stored or generated questions and evaluating their answers in real-time. 

---

## Features

- **Modular Architecture**: Organized into multiple services (e.g., interview, conversation, user, token) with clean separation of concerns.
- **Scalable Design**: Structured with Go's best practices, using models, repositories, and services for maintainability and scalability.
- **Interview Simulation**: Users can participate in mock interviews with real-time feedback.
- **Token-Based Authentication**: Secure authentication with JWT for user sessions and token refreshes.
- **Database Integration**: Fully integrated with PostgreSQL for persistent storage, using structured migrations.
- **Unit Tests**: Comprehensive testing ensures reliability of the core functionalities.

---

## Technologies Used

- **Go**: The main language used to build the backend.
- **PostgreSQL**: For relational database management.
- **JWT (JSON Web Token)**: For secure authentication and session management.
- **Chi Router**: For efficient and lightweight HTTP routing.
- **SQL Migrations**: Managed with structured SQL files to handle database versioning.

---

## Directory Structure
```
├── conversation/
│   ├── model.go
│   ├── repository.go
│   └── service.go
├── database/
│   ├── migrations/
│   │    ├── 000001_create_users_table.up.sql
│   │    ├── 000001_create_users_table.down.sql
│   │    ├── 000002_create_interviews_table.up.sql
│   │    ├── 000002_create_interviews_table.down.sql
│   │    ├── 000003_create_refresh_tokens_table.up.sql
│   │    ├── 000003_create_refresh_tokens_table.down.sql
│   │    ├── 000004_create_conversations_table.up.sql
│   │    └── 000004_create_conversations_table.down.sql
│   └── database.go
├── interview/
│   ├── model.go
│   ├── repository.go
│   ├── repository_mock.go
│   └── service.go
├── middleware/
│   └── context.go
├── token/
│   ├── model.go
│   ├── repository.go
│   ├── repository_mock.go
│   └── service.go
├── user/
│   ├── model.go
│   ├── repository.go
│   ├── repository_mock.go
│   └── service.go
├── go.mod
├── go.sum
├── handlers.go
├── handlers_test.go
├── main.go
├── README.md
```

### Key Components

- **Main App** (`main.go`): The entry point of the application. It initializes the necessary services, sets up routing, and starts the HTTP server.
  
- **Handlers**: Defined in `handlers.go`, these handle HTTP requests for various resources, interacting with services and repositories to perform necessary actions.
  
- **Models**: Define the structure of entities like `User`, `Interview`, `Conversation`, etc., and live in their respective directories (e.g., `user/model.go`).
  
- **Repositories**: Responsible for database interactions, each module has a repository (e.g., `interview/repository.go`) that abstracts SQL queries and updates.
  
- **Services**: Contain business logic (e.g., `user/service.go`) and interact with repositories for executing core functionalities.

- **Mock Repository**: Used for testing purposes (e.g., `interview/repository_mock.go`), allowing for unit tests without database dependencies.

---

## Getting Started

### Prerequisites
- **Go**: Version 1.19 or above.
- **PostgreSQL**: A running instance of PostgreSQL for local development.
- **Git**: For version control.

### Setup

1. Clone the repository:
    ```
    git clone https://github.com/your-username/interviewer-app.git
    cd interviewer-app
    ```

2. Install dependencies:
    ```
    go mod download
    ```

3. Set up your environment:
    ```
    cp .env.example .env
    ```

4. Run the SQL migrations to set up the database:
    ```
    go run main.go migrate
    ```

5. Start the application:
    ```
    go run main.go
    ```

---

## Testing

Unit tests are provided for key features of the application. You can run the tests with:
```
go test ./...
```

---

## Future Enhancements

- **Frontend Integration**: Build a frontend (in a framework like Vue or React) to make the user experience more interactive.
- **Advanced Interview Scoring**: Implement machine learning models to evaluate user answers more effectively.
- **Additional Question Types**: Support for different interview types beyond technical coding, such as system design or behavioral questions.
- **Multiple Languages**: Support for different languages beyond Python, such as Go, Javascript, C, etc . . .

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Contact

For any inquiries or suggestions, feel free to contact me:

**Michael**  
[LinkedIn](https://www.linkedin.com/in/michael-boegner-855a9741) 

