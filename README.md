Task Sync AP

The Task Sync API is a backend service designed to support an offline-first task management application. It provides a robust set of RESTful endpoints for creating, reading, updating, and deleting tasks. The core feature of this API is its asynchronous synchronization mechanism, which allows the client application to function seamlessly offline and sync data with the server once connectivity is restored.

This project is implemented in Go and utilizes SQLite for local data persistence, making it lightweight, efficient, and easy to deploy

Key Features
# Offline-First Architecture: Changes made by the client are saved locally and queued for synchronization, ensuring a smooth user experience without a constant internet connection.
# Asynchronous Synchronization: A dedicated sync queue manages all data operations (create, update, delete) and processes them in batches
# Resilient Retry Mechanism: The sync service includes a built-in retry mechanism with a configurable number of attempts to handle transient network failures gracefully.
# Conflict Resolution: A simple yet effective "last-write-wins" strategy is implemented based on timestamps to handle data conflicts during synchronization.
# RESTful Endpoints: A clean and intuitive set of API endpoints for comprehensive task management.
# Database Migrations: The database schema is managed through code-based migrations, ensuring consistency across all environments.

Prerequisites
# Go (version 1.21 or higher)
# Docker and Docker Compose
# A tool for making HTTP requests, such as curl or Postman.

Installation and Execution
# Using Docker (Recommended)
For a clean and isolated environment, it is recommended to run the application using Docker.
Build and Run with Docker Compose:
docker-compose up --build (bash)
The API will be available at http://localhost:3000

API Endpoints
# The base URL for all API endpoints is http://localhost:3000/api
Task Management
Method GET localhost:3000/api/tasks (Retrieve a list of all tasks.)
Method GET localhost:3000/api/tasks/:id (Retrieve a single task by its ID.)
Method POST localhost:3000/api/tasks (Create a new task.)
Method PUT localhost:3000/api/tasks/:id (Update an existing task.)
Method DELETE localhost:3000/api/tasks/:id (Soft delete a task.)

Synchronization
METHOD POST localhost:3000/api//sync/trigger (Trigger the synchronization process.)
Method GET localhost:3000/api//sync/status (Check the current status of the sync service.)
METHOD GET localhost:3000/api//sync/queue (View the contents of the sync queue.)

Testing
This project includes a suite of unit and integration tests to ensure the reliability and correctness of the application.
To run the tests, execute the following command from the project's task-sync-api directory:
# go test -v ./test/ (bash)
The tests utilize an in-memory SQLite database to ensure that the test environment is isolated and that tests do not interfere with each other or with any persistent data.

Architecture and Design
The application follows a standard layered architecture to separate concerns and improve maintainability:
Handlers: Responsible for parsing HTTP requests and formatting HTTP responses.
Services: Contain the core business logic of the application.
Database: Manages all interactions with the SQLite database, including migrations and data access.
Models: Define the data structures used throughout the application.

