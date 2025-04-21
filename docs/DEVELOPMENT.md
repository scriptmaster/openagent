# Development Guidelines

This document outlines the coding standards, structure, and conventions to follow when developing the OpenAgent project. Adhering to these guidelines helps maintain code clarity, consistency, and maintainability.

## General Principles

*   **Clarity:** Write code that is easy to understand and follow.
*   **Consistency:** Apply patterns and structures uniformly across the codebase.
*   **Modularity:** Organize code into logical packages based on features or domains.
*   **Testability:** Write code with testing in mind (though specific testing guidelines are TBD).

## Package Structure

The project is organized into packages based on distinct functionalities:

*   `auth`: Handles user authentication (OTP, sessions), authorization, and user management.
*   `server`: Contains the main web server setup, core HTTP handling logic not specific to other packages (like agent handlers), database initialization, and global configurations.
*   `projects`: Manages project-related data and logic.
*   `admin`: Handles administrative functions and routes.
*   `models`: Defines core data structures potentially shared across multiple packages (if any).
*   `common`: Contains truly generic utility functions reusable across different packages, primarily used to avoid import cycles (`functions.go`).
*   `main`: The entry point of the application (`main.go`), responsible for initializing the server and handling process signals.

## File Naming and Structure within Packages

Each feature package (like `auth`, `server`, `projects`) should ideally follow this structure:

*   **`routes.go`**:
    *   Responsible for registering HTTP routes for the package using `mux.HandleFunc` or similar.
    *   Should call *named* handler functions defined in `handlers.go`.
    *   Avoid defining complex request handling logic directly within inline anonymous functions. Simple middleware wrapping is acceptable.
    *   May contain package-specific middleware registration.
*   **`handlers.go`**:
    *   Contains the actual `http.HandlerFunc` implementations (e.g., `HandleLogin`, `HandleCreateProject`).
    *   Functions here take `(w http.ResponseWriter, r *http.Request)` as parameters (potentially with other dependencies like services passed in).
    *   Responsible for parsing requests, calling service/function logic, and writing HTTP responses.
*   **`types.go`**:
    *   Defines all package-specific structs and interfaces.
    *   Centralizes type definitions, making them easier to find and manage.
    *   Helps prevent import cycles if types need to be shared between different files within the same package.
*   **`service.go`**:
    *   Contains the **core business logic** and **service layer implementations** for the package.
    *   Typically involves structs (e.g., `UserService`, `ProjectService`) with methods that encapsulate major functionalities, often interacting with the database or other services.
*   **`functions.go`**:
    *   Contains **package-specific helper functions**.
    *   These are common utilities needed by multiple files *within the same package* but are not part of the core service logic and are not generic enough for the top-level `common` package (e.g., specific validation or formatting routines).

## Coding Standards

*   **Handlers:** Adhere to the standard Go `http.HandlerFunc` signature. Extract core logic into functions/methods in `service.go`.
*   **Types:** Use `types.go` for all struct/interface definitions within a package.
*   **Common Functions:** Place truly generic helper functions, needed across multiple packages (especially to avoid import cycles), in the `common/functions.go` file. Use this sparingly.
*   **Error Handling:** Functions within packages should generally return errors rather than calling `log.Fatal`. The `main` package or top-level handlers can decide how to handle errors (e.g., log and exit, return HTTP error).
*   **Exporting:** Only export functions, types, and variables (using uppercase names) that are intended for use by *other* packages. Keep internal helpers unexported (lowercase).
*   **Imports:** Keep imports organized. Use `goimports` or the editor's equivalent.
*   **Environment Variables:** Use the `common.GetEnvOrDefault` helper for safely reading environment variables with fallbacks.

## Tools and Workflow

*   **Dependencies:** Use `go mod tidy` regularly to keep `go.mod` and `go.sum` consistent.
*   **Code Verification:** Use `go vet ./...` to check for suspicious constructs and potential errors. Aim for zero `vet` errors.
* Ensure relevant tools are installed and configured. aws. eksctl. golang. dev env is setup with git repo.
* Git repo: 
