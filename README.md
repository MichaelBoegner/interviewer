# Interviewer App
**An intelligent, interactive mock interview platform powered by Go, React, and PostgreSQL, designed to help users confidently prepare for Backend Engineering interviews.**

![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker&logoColor=white)


---

## 📋 Contents

- [Overview](#-overview)
- [Learning Log](#-learning-log)
- [Recorded Demo](#-recorded-demo)
- [System Architecture](#-system-architecture)
- [Key Features](#-key-features)
- [Tech Stack](#-tech-stack)
- [API Documentation](#-api-documentation)
- [Database Schema](#-database-schema)
- [Deployment](#-deployment)
- [Local Development](#-local-development)
- [Security Implementation](#-security-implementation)
- [Performance Considerations](#-performance-considerations)
- [Testing Strategy](#-testing-strategy)
- [Development Roadmap](#-development-roadmap)
- [License](#-license)

## 🚀 Overview

Interviewer App is a robust backend service, coupled with a light demo frontend, that powers an interactive Backend Engineer interviewing experience. It leverages ChatGPT's conversational AI capabilities to create personalized, adaptive interview sessions tailored to each user's preferences.

*For a deep dive into the challenges, design decisions, and lessons learned, check out my retrospective:*  
**[Interviewer App: Lessons from Building an AI-Powered Mock Interview Platform](https://medium.com/@michaelboegner/my-experience-developing-the-mock-interview-app-interviewer-7dfc42f82ee4)**  

The application was built with a focus on:

- **Clean Architecture**: Following repository-service pattern with clear separation of concerns
- **Scalability**: Designed for horizontal scaling with stateless API design
- **Security**: Implementing industry-standard JWT authentication with refresh token rotation
- **Maintainability**: Modular code structure with reusable components

## 🧠 Learning Log

I maintain a [daily learning log](./learninglog/) as part of this project to document the challenges I face, the questions I ask, and the solutions I implement. It offers a window into how I think, debug, and grow as a backend engineer.

> 💡 If you're a hiring manager, this log is a great way to see my real-time problem-solving process and technical progression in context.

📂 [Browse the learning log →](./learninglog/)

## 🎥 Recorded Demo
- The frontend is located here: [https://github.com/michaelboegner/interviewer-ui](https://github.com/michaelboegner/interviewer-ui). 
- However, due to the costs involved with calls to OpenAI, I have opted to provide a recorded demo in lieue of open access.
- While there’s plenty of room to expand the frontend — such as adding topic listings, dashboards, and user account features — this project is primarily focused on backend engineering. My goal was to design and implement a clean, well-structured backend system with real-world patterns like service layering, token authentication, and integration with external APIs.
- Live demonstrations are available upon request. 
- [Watch the video](https://www.loom.com/share/df1cd256e2254650b0691af254747fb9?sid=0407a578-961e-4580-8425-f3066b6d183c)
[![Watch the video](assets/loom-preview.png)](https://www.loom.com/share/df1cd256e2254650b0691af254747fb9?sid=0407a578-961e-4580-8425-f3066b6d183c)

## 🏗 System Architecture

```
┌─────────────────┐      ┌──────────────────────────────────────┐      ┌───────────────┐
│                 │      │                                      │      │               │
│   Client App    │◄────►│   Go Backend API (This Repository)   │◄────►│   PostgreSQL  │
│   (React.js)    │      │                                      │      │   Database    │
│                 │      │                                      │      │               │
└─────────────────┘      └───────────────┬──────────────────────┘      └───────────────┘
                                         │
                                         │
                                         ▼
                           ┌─────────────────────────┐
                           │                         │
                           │   OpenAI API (ChatGPT)  │
                           │                         │
                           └─────────────────────────┘
```

The backend is structured using a layered architecture:

- **Handler Layer**: Request validation and response formation
- **Service Layer**: Business logic encapsulation
- **Repository Layer**: Data access and persistence
- **Middleware**: Cross-cutting concerns (authentication, logging, etc.)

## 🎯 Key Features

- **User Management**: Secure user registration, authentication, and profile management
- **JWT-based Authentication**: Access tokens with configurable expiration and refresh token rotation
- **Structured Mock Interviews**: Dynamic interview generation
- **Conversational AI Integration**: Seamless integration with OpenAI's GPT model
- **Persistent Data Storage**: Complete interview history stored for future review
- **RESTful API Design**: Consistent and predictable API endpoints
- **Middleware Pipeline**: Extensible middleware for request processing
- **Environment-based Configuration**: Flexible configuration for different deployment environments

## 🛠️ Tech Stack

| Component             | Technology                                       |
|-----------------------|--------------------------------------------------|
| **Backend Language**  | Go (Golang) 1.20+                               |
| **Database**          | PostgreSQL 15+                                   |
| **Authentication**    | JWT-based authentication (access & refresh tokens) |
| **AI Integration**    | OpenAI GPT API (4.0)                             |
| **Testing**           | Go table-driven tests (unit + integration)       |
| **Containerization**  | Docker                                           |
| **Deployment**        | Fly.io                                           |
| **Database Hosting**  | Supabase (PostgreSQL)                            |
| **Version Control**   | Git                                              |

## 📘 API Documentation

### Authentication Flow

```
┌─────────┐                                  ┌─────────┐                       ┌─────────┐
│         │                                  │         │                       │         │
│ Client  │                                  │ Server  │                       │ Database│
│         │                                  │         │                       │         │
└────┬────┘                                  └────┬────┘                       └────┬────┘
     │                                            │                                 │
     │ POST /api/users (Register)                 │                                 │
     │───────────────────────────────────────────►│                                 │
     │                                            │                                 │
     │                                            │ Store User                      │
     │                                            │────────────────────────────────►│
     │                                            │                                 │
     │ 201 Created                                │                                 │
     │◄───────────────────────────────────────────│                                 │
     │                                            │                                 │
     │ POST /api/auth/login                       │                                 │
     │───────────────────────────────────────────►│                                 │
     │                                            │ Verify Credentials              │
     │                                            │────────────────────────────────►│
     │                                            │                                 │
     │ 200 OK (Access Token + Refresh Token)      │                                 │
     │◄───────────────────────────────────────────│                                 │
     │                                            │                                 │
     │ Request with Access Token                  │                                 │
     │───────────────────────────────────────────►│                                 │
     │                                            │ Validate Token                  │
     │                                            │                                 │
     │ Response                                   │                                 │
     │◄───────────────────────────────────────────│                                 │
     │                                            │                                 │
     │ POST /api/auth/token (Refresh)             │                                 │
     │───────────────────────────────────────────►│                                 │
     │                                            │ Verify Refresh Token            │
     │                                            │────────────────────────────────►│
     │                                            │                                 │
     │ 200 OK (New Access Token + Refresh Token)  │                                 │
     │◄───────────────────────────────────────────│                                 │
     │                                            │                                 │
```

### API Endpoints

#### User Management
- `POST /api/users` - Register a new user
- `GET /api/users/{id}` - Get user profile

#### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/token` - Refresh access token

#### Interviews
- `POST /api/interviews` - Create a new interview

#### Conversations
- `POST /api/conversations/create` - Create a new conversation
- `POST /api/conversations/append` - Append an existing conversation

## 🗄 Database Schema

```
┌─────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│     users       │       │    interviews    │       │  conversations   │
├─────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id              │       │ id               │       │ id               │
│ username        │       │ user_id          │◄──────┤ interview_id     │
│ email           │       │ length           │       │ current_topic    │
│ password        │       │ number_questions │       │ current_subtopic │
│ created_at      │       │ difficulty       │       │ current_question_│
│ updated_at      │       │ status           │       │ created_at       │
└────────┬────────┘       │ score            │       │ updated_at       │
         │                │ language         │       └────────┬─────────┘
         │                │ prompt           │                │
         │                │ first_question   │                │
         │                │ subtopic         │                │
         │                │ created_at       │                │
         │                │ updated_at       │                │
         │                └──────────────────┘                │
         │                                                    │
         │                                                    │
┌────────▼────────┐       ┌──────────────────┐       ┌────────▼─────────┐
│ refresh_tokens  │       │    questions     │       │     messages     │
├─────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id              │       │ id               │       │ id               │
│ user_id         │       │ conversation_id  │◄──────┤ conversation_id  │
│ refresh_token   │       │ topic_id         │       │ topic_id         │
│ expires_at      │       │ question_number  │       │ question_number  │
│ created_at      │       │ prompt           │       │ author           │
│ updated_at      │       │ created_at       │       │ content          │
└─────────────────┘       └──────────────────┘       │ created_at       │
                                                     └──────────────────┘
```

## 📦 Deployment

This application is currently deployed on Fly.io with a PostgreSQL database hosted on Supabase.

### Deployment Stack
- **Application Hosting**: Fly.io (containerized deployment)
- **Database**: Supabase PostgreSQL
- **Environment Variables**: Managed through Fly.io secrets

### Deployment Process
1. Build Docker container using the included Dockerfile
2. Push to Fly.io platform
3. Configure environment variables for production

```bash
# Deploy to Fly.io
fly launch
fly secrets set JWT_SECRET=your-secret DATABASE_URL=your-supabase-url OPENAI_API_KEY=your-api-key
fly deploy
```

## 🔧 Local Development

### Prerequisites
- Go 1.20+
- PostgreSQL 15+
- Docker (optional, for containerized development)

### Setup Instructions

1. Clone the repository
   ```bash
   git clone https://github.com/yourusername/interviewer.git
   cd interviewer
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Set up environment variables
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Create the database
   ```bash
   # Create PostgreSQL database named 'interviewerio'
   createdb interviewerio
   
   # Apply migrations
   # Either manually execute SQL files in database/migrations or use a migration tool
   ```

5. Run the application
   ```bash
   go run main.go
   ```

### Using Docker
```bash
docker build -t interviewer .
docker run -p 8080:8080 --env-file .env interviewer
```

## 🔒 Security Implementation

- **Password Hashing**: Passwords are securely hashed and never stored in plaintext
- **JWT Authentication**: Short-lived access tokens with refresh token rotation
- **Prepared Statements**: All database queries use prepared statements to prevent SQL injection
- **CORS Configuration**: Configured to restrict origins in production environments
- **Environment Variables**: Sensitive configuration stored in environment variables

## ⚡ Performance Considerations

- **Stateless Design**: The API is designed to be stateless, allowing for horizontal scaling

## ✅ Testing Strategy

The app is tested with both **unit tests** (mocked repositories) and **integration tests** (real PostgreSQL + full HTTP stack).

- **Unit Tests**: Every core domain (`user`, `token`, `interview`, `conversation`) is covered using table-driven tests and a mock repository layer. This verifies business logic independently from the database.
- **Integration Tests**: `handlers/handlers_test.go` validates full end-to-end request flows using a Dockerized PostgreSQL test database. The `Makefile` automates setup, teardown, and migration steps to simulate a production-like environment.

Tests are run consistently during development to ensure correctness, stability, and maintainability. CI is coming next (see Development Roadmap below)

## 🛣️ Development Roadmap

### Current Focus
- Implementation of CI/CD pipeline

### Upcoming Improvements
- Enhancing error handling and recovery mechanisms
- Structured logging for better observability
- Input validation
- Implement a single active session policy
- Supporting multiple conversation tracks within an interview
- Adding detailed analytics for interview performance
- Optimizing database queries
- Refining API documentation
- Interview preferences (language, difficulty, duration, etc . . .)

## 📜 License

Copyright (c) 2024 Michael Boegner

This source code is proprietary. 
All rights reserved. No part of this code may be reproduced, distributed, or used 
without explicit permission from the author.

---

**Created by Michael Boegner** - [GitHub Profile](https://github.com/michaelboegner)



