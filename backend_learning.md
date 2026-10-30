# Complete Process Implementation & Backend Engineering Guide: Update & Delete Student Methods

This guide provides an end-to-end walkthrough of implementing the **Update (`PUT`)** and **Delete (`DELETE`)** methods in the **Go Student Management REST API**. It covers the architectural design, core backend engineering concepts, step-by-step code implementation, error handling strategies, and comprehensive API testing instructions.

---

## Table of Contents
1. [Architectural Overview & Request Lifecycle](#1-architectural-overview--request-lifecycle)
2. [Core Backend Concepts: UPDATE & DELETE in REST APIs](#2-core-backend-concepts-update--delete-in-rest-apis)
3. [Implementation Process Step-by-Step](#3-implementation-process-step-by-step)
   - [Step 1: Domain Entity & Struct Tags (`internal/types/types.go`)](#step-1-domain-entity--struct-tags-internaltypestypesgo)
   - [Step 2: Defining the Contract (`internal/storage/storage.go`)](#step-2-defining-the-contract-internalstoragestoragego)
   - [Step 3: Database Implementation with SQLite (`internal/storage/sqlite/sqlite.go`)](#step-3-database-implementation-with-sqlite-internalstoragesqlitesqlitego)
   - [Step 4: HTTP Handlers Layer (`internal/http/handlers/student/student.go`)](#step-4-http-handlers-layer-internalhttphandlersstudentstudentgo)
   - [Step 5: Route Registration & Server Wiring (`cmd/students-api/main.go`)](#step-5-route-registration--server-wiring-cmdstudents-apimaingo)
4. [Critical Backend Learnings & Deep Dives](#4-critical-backend-learnings--deep-dives)
   - [Why `RowsAffected()` is Essential for UPDATE and DELETE](#why-rowsaffected-is-essential-for-update-and-delete)
   - [Prepared Statements & SQL Injection Defense](#prepared-statements--sql-injection-defense)
   - [Input Validation Before Persistence (Fail-Fast Pattern)](#input-validation-before-persistence-fail-fast-pattern)
   - [HTTP Status Codes Mapping (400 vs 404 vs 500)](#http-status-codes-mapping-400-vs-404-vs-500)
   - [Structured Logging with `log/slog`](#structured-logging-with-logslog)
5. [Complete Hands-on Testing Guide](#5-complete-hands-on-testing-guide)
   - [Running the Server](#running-the-server)
   - [cURL Commands](#curl-commands)
   - [PowerShell Commands (Windows Native)](#powershell-commands-windows-native)
6. [Automated Testing with `httptest`](#6-automated-testing-with-httptest)

---

## 1. Architectural Overview & Request Lifecycle

This application is structured according to **Clean Architecture / Layered Architecture** principles. In Go backends, separating concerns into distinct layers ensures loose coupling, testability, and maintainability.

```
Client (cURL / Postman / Browser)
   │
   ▼
[1] HTTP Router (`net/http.ServeMux` in `cmd/students-api/main.go`)
   │  - Matches Method + URL Pattern ("PUT /api/students/{id}", "DELETE /api/students/{id}")
   │
   ▼
[2] HTTP Handler Layer (`internal/http/handlers/student/student.go`)
   │  - Extracts path parameters (`r.PathValue("id")`)
   │  - Parses & decodes JSON request payload
   │  - Validates request fields using `validator/v10`
   │  - Logs structured context using `log/slog`
   │  - Calls Storage interface
   │  - Formats & writes JSON HTTP responses + status codes
   │
   ▼
[3] Storage Abstraction Layer (`internal/storage/storage.go`)
   │  - Declares Go Interface (The Contract)
   │  - Inversion of Control: Handlers depend on abstraction, not database driver
   │
   ▼
[4] Database Implementation Layer (`internal/storage/sqlite/sqlite.go`)
   │  - Prepares SQL statements (`db.Prepare`)
   │  - Executes queries safely with parameter binding
   │  - Inspects `RowsAffected()` to verify record existence
   │
   ▼
[5] Database Engine (SQLite DB file: `storage/storage.db`)
```

### Why Layered Architecture Matters in Backend Development:
- **Testability**: The handler only requires `storage.Storage` interface. In unit tests, we can supply an in-memory SQLite implementation or a mock struct without touching a production database.
- **Portability**: If you decide to switch SQLite to PostgreSQL or MySQL tomorrow, only the `internal/storage/` package changes; the HTTP handlers and routing logic remain 100% untouched.

---

## 2. Core Backend Concepts: UPDATE & DELETE in REST APIs

### 2.1 The HTTP Methods
| HTTP Verb | Operation | Idempotent? | Safe? | Typical Status Codes |
|---|---|---|---|---|
| `POST` | Create | No | No | `201 Created`, `400 Bad Request` |
| `GET` | Read | Yes | Yes | `200 OK`, `404 Not Found` |
| `PUT` | Full Update / Replace | Yes | No | `200 OK`, `400 Bad Request`, `404 Not Found` |
| `PATCH` | Partial Update | No (typically) | No | `200 OK`, `400 Bad Request`, `404 Not Found` |
| `DELETE` | Delete | Yes | No | `200 OK`, `204 No Content`, `404 Not Found` |

#### Understanding Idempotency
An HTTP method is **idempotent** if making the same request multiple times produces the identical side-effect on the server as making it once.
- `PUT`: If you update student #1 with `{name: "Alice", email: "alice@example.com", age: 22}` 10 times, the database row ends up with the exact same values every time.
- `DELETE`: If you delete student #1, the record is removed. If you call it again, student #1 remains absent (and the server returns 404).
- `POST`: NOT idempotent. Submitting POST 5 times creates 5 separate database records.

---

## 3. Implementation Process Step-by-Step

### Step 1: Domain Entity & Struct Tags (`internal/types/types.go`)

The domain model represents the core business object. We added the `Id` field with `json:"id,omitempty"` so that queries returning students can include their primary key, while incoming creation bodies do not fail validation.

```go
package types

type Student struct {
    Id    int64  `json:"id,omitempty"`
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,numeric,min=18"`
}
```

#### Key Backend Concepts:
1. **JSON Struct Tags (`json:"..."`)**: Tells Go's `encoding/json` encoder and decoder what keys to look for in JSON payloads. `omitempty` ensures that when serializing an entity without an ID (zero value), it does not emit `"id": 0`.
2. **Validation Tags (`validate:"..."`)**: Used by `go-playground/validator/v10`:
   - `required`: Field must not be empty or zero.
   - `email`: Enforces standard email syntax.
   - `min=18`: Enforces business rule (students must be at least 18).

---

### Step 2: Defining the Contract (`internal/storage/storage.go`)

We define the interface containing the new operations:

```go
package storage

import "github.com/kartik23-2/Student-API/internal/types"

type Storage interface {
    CreateStudent(name string, email string, age int) (int64, error)
    GetStudentById(id int64) (types.Student, error)
    GetAll() ([]types.Student, error)
    UpdateStudent(id int64, name string, email string, age int) error
    DeleteStudent(id int64) error
}
```

#### Key Backend Concept: Interface Segregation & Dependency Inversion
The HTTP layer does not import or know about SQLite. It only knows about `storage.Storage`. This is the **Dependency Inversion Principle (DIP)** in SOLID software design.

---

### Step 3: Database Implementation with SQLite (`internal/storage/sqlite/sqlite.go`)

Now we implement `UpdateStudent` and `DeleteStudent` on `*Sqlite`:

```go
// UpdateStudent modifies an existing student's record by ID.
func (s *Sqlite) UpdateStudent(id int64, name string, email string, age int) error {
    stmt, err := s.Db.Prepare("UPDATE students SET name = $1, email = $2, age = $3 WHERE id = $4")
    if err != nil {
        return err
    }
    defer stmt.Close()

    result, err := stmt.Exec(name, email, age, id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return fmt.Errorf("student not found with id: %d", id)
    }

    return nil
}

// DeleteStudent removes a student record from the database by ID.
func (s *Sqlite) DeleteStudent(id int64) error {
    stmt, err := s.Db.Prepare("DELETE FROM students WHERE id = $1")
    if err != nil {
        return err
    }
    defer stmt.Close()

    result, err := stmt.Exec(id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return fmt.Errorf("student not found with id: %d", id)
    }

    return nil
}
```

We also updated `GetStudentById` and `GetAll` to include `id` in the `SELECT` query:
```go
// In GetStudentById:
stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students WHERE id = $1")
...
err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age)

// In GetAll:
stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students")
...
err := rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age)
```

---

### Step 4: HTTP Handlers Layer (`internal/http/handlers/student/student.go`)

We implement the HTTP handler closures for `Update` and `Delete`:

```go
// Update handles PUT /api/students/{id}
func Update(storage storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        slog.Info("update student by id", slog.String("id", id))

        intId, err := strconv.ParseInt(id, 10, 64)
        if err != nil {
            response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid student id: %s", id)))
            return
        }

        var student types.Student
        err = json.NewDecoder(r.Body).Decode(&student)
        if errors.Is(err, io.EOF) {
            response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
            return
        }
        if err != nil {
            response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
            return
        }

        // Validation request
        if err := validator.New().Struct(student); err != nil {
            response.WriteJson(w, http.StatusBadRequest, response.ValidationError(err.(validator.ValidationErrors)))
            return
        }

        err = storage.UpdateStudent(intId, student.Name, student.Email, student.Age)
        if err != nil {
            slog.Error("error in updating student", slog.String("error", err.Error()))
            if strings.Contains(err.Error(), "not found") {
                response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
                return
            }
            response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
            return
        }

        student.Id = intId
        slog.Info("student updated successfully", slog.Int64("id", intId))
        response.WriteJson(w, http.StatusOK, map[string]interface{}{
            "message": "student updated successfully",
            "student": student,
        })
    }
}

// Delete handles DELETE /api/students/{id}
func Delete(storage storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        slog.Info("delete student by id", slog.String("id", id))

        intId, err := strconv.ParseInt(id, 10, 64)
        if err != nil {
            response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid student id: %s", id)))
            return
        }

        err = storage.DeleteStudent(intId)
        if err != nil {
            slog.Error("error in deleting student", slog.String("error", err.Error()))
            if strings.Contains(err.Error(), "not found") {
                response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
                return
            }
            response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
            return
        }

        slog.Info("student deleted successfully", slog.Int64("id", intId))
        response.WriteJson(w, http.StatusOK, map[string]interface{}{
            "message": "student deleted successfully",
            "id":      intId,
        })
    }
}
```

---

### Step 5: Route Registration & Server Wiring (`cmd/students-api/main.go`)

In Go 1.22+, `http.ServeMux` natively supports HTTP method matching and path parameter patterns:

```go
// setup router
router := http.NewServeMux()
router.HandleFunc("POST /api/students", student.New(storage))
router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
router.HandleFunc("GET /api/students", student.GetAll(storage))
router.HandleFunc("PUT /api/students/{id}", student.Update(storage))
router.HandleFunc("DELETE /api/students/{id}", student.Delete(storage))
```

We also refined the graceful shutdown block:
```go
go func() {
    err := server.ListenAndServe()
    if err != nil && err != http.ErrServerClosed {
        log.Fatalf("server error: %s", err.Error())
    }
}()
```
When `server.Shutdown` is invoked on `SIGINT` (Ctrl+C), `ListenAndServe` returns `http.ErrServerClosed`. Checking this prevents false fatal crash logs during clean shutdown.

---

## 4. Critical Backend Learnings & Deep Dives

### Why `RowsAffected()` is Essential for UPDATE and DELETE
One of the most common pitfalls in backend development is assuming that if `stmt.Exec()` returns `err == nil`, the record was updated or deleted.

In SQL databases (SQLite, MySQL, Postgres):
```sql
UPDATE students SET name = 'John' WHERE id = 999999;
DELETE FROM students WHERE id = 999999;
```
If ID `999999` does not exist, the SQL query executes **successfully with 0 errors** because the query itself is syntactically and logically valid; it simply matched zero rows!

```go
result, err := stmt.Exec(id)
if err != nil {
    return err // Only triggered if database connection dropped, syntax error, etc.
}

rowsAffected, err := result.RowsAffected()
if rowsAffected == 0 {
    // Record didn't exist! We MUST return an error here so handler can return 404 Not Found.
    return fmt.Errorf("student not found with id: %d", id)
}
```
Without checking `RowsAffected()`, an API would mistakenly send `200 OK: "student deleted successfully"` for non-existent IDs.

---

### Prepared Statements & SQL Injection Defense
Notice we always use:
```go
stmt, err := s.Db.Prepare("UPDATE students SET name = $1, email = $2, age = $3 WHERE id = $4")
...
result, err := stmt.Exec(name, email, age, id)
```
**NEVER** format SQL strings with user input:
```go
// DANGEROUS! SQL Injection vulnerability!
query := fmt.Sprintf("UPDATE students SET name = '%s' WHERE id = %d", student.Name, id)
s.Db.Exec(query)
```
Prepared statements compile the query structure first, and then send the parameters separately in a binary format. The database engine never interprets parameter contents as executable SQL code.

---

### Input Validation Before Persistence (Fail-Fast Pattern)
Always validate inputs before calling the database:
1. **URL Parameter**: `strconv.ParseInt(id, 10, 64)` ensures non-numeric IDs like `/api/students/abc` are rejected immediately with `400 Bad Request`.
2. **JSON Stream**: `json.NewDecoder(r.Body).Decode(&student)` detects malformed JSON and empty bodies (`io.EOF`).
3. **Domain Validation**: `validator.New().Struct(student)` checks emails, age limits, and required strings.

**Benefit**: Prevents wasting database connections and CPU cycles on invalid requests.

---

### HTTP Status Codes Mapping (400 vs 404 vs 500)
- **`400 Bad Request`**: The client sent invalid data (bad ID syntax, missing required fields, age under 18, empty body). The client must fix their request.
- **`404 Not Found`**: The request is formatted properly, but the resource identified by the ID does not exist in the database.
- **`500 Internal Server Error`**: An unexpected failure occurred on the server (e.g., database connection loss, disk I/O error).
- **`200 OK`**: The operation succeeded and the response contains relevant data or confirmation.

---

### Structured Logging with `log/slog`
Go 1.21 introduced the `log/slog` package into the standard library.
```go
slog.Info("student updated successfully", slog.Int64("id", intId))
slog.Error("error in updating student", slog.String("error", err.Error()))
```
Benefits over `log.Println` or `fmt.Printf`:
- Produces key-value pairs (`id=1`, `error="..."`).
- Can easily be configured to output machine-readable JSON in production log aggregators (Datadog, CloudWatch, Grafana Loki).

---

## 5. Complete Hands-on Testing Guide

### Running the Server
Ensure the config file exists at `config/local.yaml`:
```bash
go run cmd/students-api/main.go --config config/local.yaml
```
Output:
```
2026/10/30 21:10:00 INFO storage initialised env=dev version=1.0.0
2026/10/30 21:10:00 INFO server started address=localhost:8080
```

---

### cURL Commands

#### 1. Create a Student (POST)
```bash
curl -X POST http://localhost:8080/api/students \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Smith", "email": "alice@example.com", "age": 21}'
```
**Response (201 Created):**
```json
{
  "id": 1
}
```

#### 2. Get All Students (GET)
```bash
curl -X GET http://localhost:8080/api/students
```
**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "Alice Smith",
    "email": "alice@example.com",
    "age": 21
  }
]
```

#### 3. Get Student by ID (GET)
```bash
curl -X GET http://localhost:8080/api/students/1
```
**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Alice Smith",
  "email": "alice@example.com",
  "age": 21
}
```

#### 4. Update Student (PUT) - Valid
```bash
curl -X PUT http://localhost:8080/api/students/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Johnson", "email": "alice.johnson@example.com", "age": 22}'
```
**Response (200 OK):**
```json
{
  "message": "student updated successfully",
  "student": {
    "id": 1,
    "name": "Alice Johnson",
    "email": "alice.johnson@example.com",
    "age": 22
  }
}
```

#### 5. Update Student (PUT) - Validation Failure (Age < 18)
```bash
curl -X PUT http://localhost:8080/api/students/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Johnson", "email": "alice@example.com", "age": 16}'
```
**Response (400 Bad Request):**
```json
{
  "status": 400,
  "error": "Age is too short"
}
```

#### 6. Update Student (PUT) - Non-Existent ID
```bash
curl -X PUT http://localhost:8080/api/students/9999 \
  -H "Content-Type: application/json" \
  -d '{"name": "Non Existent", "email": "ghost@example.com", "age": 25}'
```
**Response (404 Not Found):**
```json
{
  "status": 400,
  "error": "student not found with id: 9999"
}
```

#### 7. Delete Student (DELETE) - Valid
```bash
curl -X DELETE http://localhost:8080/api/students/1
```
**Response (200 OK):**
```json
{
  "id": 1,
  "message": "student deleted successfully"
}
```

#### 8. Delete Student (DELETE) - Already Deleted / Non-Existent
```bash
curl -X DELETE http://localhost:8080/api/students/1
```
**Response (404 Not Found):**
```json
{
  "status": 400,
  "error": "student not found with id: 1"
}
```

---

### PowerShell Commands (Windows Native)

If you are using Windows PowerShell:

```powershell
# 1. Create a Student
Invoke-RestMethod -Uri "http://localhost:8080/api/students" -Method POST -ContentType "application/json" -Body '{"name":"John Doe","email":"john@example.com","age":20}'

# 2. Get All Students
Invoke-RestMethod -Uri "http://localhost:8080/api/students" -Method GET

# 3. Get Student by ID
Invoke-RestMethod -Uri "http://localhost:8080/api/students/1" -Method GET

# 4. Update Student
Invoke-RestMethod -Uri "http://localhost:8080/api/students/1" -Method PUT -ContentType "application/json" -Body '{"name":"John Smith","email":"john.smith@example.com","age":21}'

# 5. Delete Student
Invoke-RestMethod -Uri "http://localhost:8080/api/students/1" -Method DELETE
```

---

## 6. Automated Testing with `httptest`

To guarantee regression-free code, we implemented automated unit and integration tests in `internal/http/handlers/student/student_test.go` using an in-memory SQLite database (`file::memory:?cache=shared`).

Run the test suite at any time:
```bash
go test -v ./...
```

Expected result:
```
=== RUN   TestStudentCRUDLifecycle
--- PASS: TestStudentCRUDLifecycle (0.01s)
PASS
ok      github.com/kartik23-2/Student-API/internal/http/handlers/student    0.055s
```

All 8 stages of the lifecycle (Create -> Get -> Update -> Verify Update -> Reject Invalid Update -> Delete -> Verify Deletion -> Reject Non-Existent Deletion) are automatically executed and verified!

