# Student Management System API

**Version:** 1.0.0  
**Language:** Python  
**Framework:** FastAPI  
**Database:** in-memory (SQLite-compatible)

## 🚀 Quick Start

### Prerequisites
- Python 3.7+

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd student-management-api

# Create a virtual environment
python -m venv venv

# Activate the virtual environment
# Windows
venv\Scripts\activate

# Linux/Mac
source venv/bin/activate

# Install dependencies
pip install fastapi uvicorn
```

### Run the server
```bash
uvicorn main:app --reload
```

The API will be available at `http://localhost:8000`

## 📖 API Documentation

The API provides the following endpoints:

### 📂swagger-ui  
Interactive API documentation
- **URL:** `http://localhost:8000/docs`
- **Access:** Open this URL in your browser to test endpoints interactively

### 📄redoc
Alternative API documentation with endpoint details
- **URL:** `http://localhost:8000/redoc`

## 📋 Endpoints Reference

### ✅ All Students  
Get list of all students
```
GET /students
```

### ✅ Add Student  
Add a new student
```
POST /students

Request Body:
{
  "name": "string",
  "age": "integer",
  "standard": "string",
  "subjects": ["string"]
}
```

### ✅ Get Student by ID  
Get a single student by ID
```
GET /students/{student_id}
```

### ✅ Update Student  
Update an existing student
```
PUT /students/{student_id}

Request Body:
{
  "name": "string (optional)",
  "age": "integer (optional)",
  "standard": "string (optional)",
  "subjects": ["string"] (optional)
}
```

### ✅ Delete Student  
Delete a student by ID
```
DELETE /students/{student_id}
```

## 📝 Usage Examples

### Add a new student
```bash
curl -X POST "http://localhost:8000/students" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "John Doe",
       "age": 16,
       "standard": "10th",
       "subjects": ["Math", "Science", "English"]
     }'
```

### Get all students
```bash
curl "http://localhost:8000/students"
```

### Get a specific student
```bash
curl "http://localhost:8000/students/1"
```

### Update a student
```bash
curl -X PUT "http://localhost:8000/students/1" \
     -H "Content-Type: application/json" \
     -d '{
       "age": 17,
       "subjects": ["Math", "Science", "English", "Physics"]
     }'
```

### Delete a student
```bash
curl -X DELETE "http://localhost:8000/students/1"
```

## 🏗️ Architecture

```
┌────────────────────────────┐
│        FastAPI             │
│      (Web Framework)       │
└─────────────┬──────────────┘
              │
    ┌─────────┴─────────────┐
    │  StudentService         │
    │  (Business Logic)       │
    └─────────────┬─────────────┘
                  │
    ┌─────────────┴─────────────┐
    │   StudentRepository         │
    │  (In-memory Storage)      │
    └────────────────────────────┘
```

## 📂 Project Structure

```
student-management-api/
├── app/
│   ├── main.py           # FastAPI app entry point
│   ├── database.py       # Database configuration
│   ├── models.py         # Data models
│   ├── schemas.py        # Request/response schemas
│   ├── services.py       # Business logic
│   ├── repositories.py   # Data access layer
│   └── main.py           # FastAPI app entry point
├── venv/                 # Virtual environment
├── requirements.txt      # Dependencies
└── README.md             # Project documentation
```

## 🧩 API Error Responses

### ❌ 404 Not Found
```json
{
  "detail": "Student not found"
}
```

### ❌ 422 Validation Error
```json
{
  "detail": [
    {
      "type": "string_type",
      "loc": ["string"],
      "msg": "string",
      "input": {}
    }
  ]
}
```

## 🧪 Testing

### Manual Testing
Use the interactive Swagger UI at `http://localhost:8000/docs` to test endpoints.

### Automated Tests
```bash
# Run tests
python -m pytest tests/
```

## 🔧 Configuration

Edit `app/config.py` to modify database settings or other configurations.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Team

- **Your Name** - [Your GitHub](https://github.com/your-username)
