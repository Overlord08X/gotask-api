# GoTask - Sistem Manajemen Tugas

Sistem manajemen tugas berbasis web backend yang dibangun dengan Go menggunakan Clean Architecture.

## Tech Stack
- **Go** - Bahasa pemrograman
- **Gin** - HTTP framework
- **MongoDB Go Driver** - Database driver
- **MongoDB** - Database
- **JWT** - Autentikasi
- **UUID** - Primary key

## Struktur Project (Clean Architecture)
```
├── entity/          # Model/entitas database
├── repository/      # Interface & implementasi akses data
├── service/         # Business logic
├── dto/             # Data Transfer Object (request/response)
├── config/          # Konfigurasi database
├── middleware/      # JWT auth middleware
├── docs/            # Dokumentasi rancangan sistem
└── main.go          # Entry point & routing
```

## Setup (Docker)
```bash
docker compose up --build
```
Server jalan di `http://localhost:8080`, PostgreSQL di port `5432`.

### Tanpa Docker
1. Copy `.env.example` ke `.env` dan sesuaikan konfigurasi
2. Buat database PostgreSQL `gotask`
3. Jalankan:
```bash
go mod tidy
go run main.go
```

## API Endpoints

### Public
| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/register` | Register user baru |
| POST | `/api/login` | Login |

### User (Auth required)
| Method | Endpoint | Role | Deskripsi |
|--------|----------|------|-----------|
| GET | `/api/users` | Admin | List semua user |
| GET | `/api/users/:id` | All | Detail user |
| PUT | `/api/users/:id` | All | Update user |
| DELETE | `/api/users/:id` | Admin | Hapus user |

### Project (Auth required)
| Method | Endpoint | Role | Deskripsi |
|--------|----------|------|-----------|
| POST | `/api/projects` | Admin, Manager | Buat project |
| GET | `/api/projects` | All | List project |
| GET | `/api/projects/:id` | All | Detail project |
| PUT | `/api/projects/:id` | Admin, Manager | Update project |
| DELETE | `/api/projects/:id` | Admin, Manager | Hapus project |
| POST | `/api/projects/:id/members` | Admin, Manager | Tambah member |
| GET | `/api/projects/:id/members` | All | List member |
| DELETE | `/api/projects/:id/members/:user_id` | Admin, Manager | Hapus member |

### Task (Auth required)
| Method | Endpoint | Role | Deskripsi |
|--------|----------|------|-----------|
| POST | `/api/projects/:id/tasks` | Admin, Manager | Buat task |
| GET | `/api/projects/:id/tasks` | All | List task per project |
| GET | `/api/tasks/me` | All | Task saya |
| GET | `/api/tasks/:id` | All | Detail task |
| PUT | `/api/tasks/:id` | Admin, Manager | Update task |
| PATCH | `/api/tasks/:id/status` | All | Update status task |
| DELETE | `/api/tasks/:id` | Admin, Manager | Hapus task |

### Dashboard (Auth required)
| Method | Endpoint | Role | Deskripsi |
|--------|----------|------|-----------|
| GET | `/api/dashboard` | All | Ringkasan project & task |
