# GoTask - Sistem Manajemen Tugas

## 1. Deskripsi Sistem

Sistem manajemen tugas berbasis web backend yang memungkinkan pengelolaan proyek dan tugas secara kolaboratif. Dibangun dengan Go menggunakan Clean Architecture.

---

## 2. Tech Stack

| Komponen | Teknologi |
|---|---|
| Bahasa | Go |
| HTTP Framework | Gin |
| Database Driver | MongoDB Go Driver (`go.mongodb.org/mongo-driver`) |
| Database | MongoDB |
| Autentikasi | JWT (JSON Web Token) |
| ID Generator | ObjectID (MongoDB `_id`, disimpan sebagai string) |
| Arsitektur | Clean Architecture (Entity → Repository → Service → DTO) |

---

## 3. Fitur-Fitur

### 3.1 Autentikasi
- Register akun baru
- Login dengan email & password
- Proteksi endpoint menggunakan JWT

### 3.2 Manajemen User
- CRUD data user
- Lihat & update profile

### 3.3 Manajemen Project
- CRUD project
- Assign member ke project

### 3.4 Manajemen Task
- CRUD task dalam project
- Assign task ke member project
- Set priority (low, medium, high)
- Set deadline

### 3.5 Status Tracking
- Update status task: `todo` → `in_progress` → `done`

### 3.6 Dashboard
- Ringkasan jumlah project
- Jumlah task per status (todo, in_progress, done)

---

## 4. Role pada Sistem

| Role | Hak Akses |
|---|---|
| **Admin** | Kelola semua user, project, dan task |
| **Manager** | Buat & kelola project sendiri, assign member ke project, buat task |
| **Member** | Lihat project yang di-assign, update status task yang di-assign |

---

## 5. Rancangan Database (MongoDB Collections)

> **Catatan:** MongoDB adalah database NoSQL berbasis dokumen. Tidak ada foreign key constraint seperti di database relasional. Relasi antar dokumen dilakukan dengan menyimpan ID dokumen lain sebagai reference. Validasi referensi dilakukan di level aplikasi (service layer).

### 5.1 Collection `users`

| Field | Tipe | Keterangan |
|---|---|---|
| _id | ObjectID | Primary Key (otomatis dari MongoDB) |
| name | string | Nama lengkap |
| email | string | Unique (gunakan unique index) |
| password | string | Hashed password |
| role | string | `admin` / `manager` / `member` |
| created_at | datetime | Waktu pembuatan |
| updated_at | datetime | Waktu update terakhir |

### 5.2 Collection `projects`

| Field | Tipe | Keterangan |
|---|---|---|
| _id | ObjectID | Primary Key (otomatis dari MongoDB) |
| name | string | Nama project |
| description | string | Deskripsi project |
| created_by | ObjectID (string) | Reference → `users._id` |
| created_at | datetime | Waktu pembuatan |
| updated_at | datetime | Waktu update terakhir |

### 5.3 Collection `project_members`

| Field | Tipe | Keterangan |
|---|---|---|
| _id | ObjectID | Primary Key (otomatis dari MongoDB) |
| project_id | ObjectID (string) | Reference → `projects._id` |
| user_id | ObjectID (string) | Reference → `users._id` |
| joined_at | datetime | Waktu bergabung |

> **Index:** Buat compound unique index pada `(project_id, user_id)` untuk mencegah duplikasi member.

### 5.4 Collection `tasks`

| Field | Tipe | Keterangan |
|---|---|---|
| _id | ObjectID | Primary Key (otomatis dari MongoDB) |
| project_id | ObjectID (string) | Reference → `projects._id` |
| title | string | Judul task |
| description | string | Deskripsi task |
| status | string | `todo` / `in_progress` / `done` |
| priority | string | `low` / `medium` / `high` |
| assigned_to | ObjectID (string) | Reference → `users._id` (nullable) |
| deadline | datetime | Batas waktu (nullable) |
| created_by | ObjectID (string) | Reference → `users._id` |
| created_at | datetime | Waktu pembuatan |
| updated_at | datetime | Waktu update terakhir |

### 5.5 Relasi Antar Collection

> Di MongoDB, relasi tidak menggunakan foreign key constraint. Sebagai gantinya, field reference menyimpan `_id` dari dokumen di collection lain. Validasi keberadaan dokumen yang direferensikan dilakukan di level aplikasi.

```
users 1───N projects          (created_by → menyimpan users._id)
users M───N projects          (via collection project_members)
projects 1───N tasks          (project_id → menyimpan projects._id)
users 1───N tasks             (assigned_to → menyimpan users._id)
users 1───N tasks             (created_by → menyimpan users._id)
```

**Diagram:**

```
┌──────────┐       ┌──────────────────┐       ┌──────────────┐
│  users   │1─────N│ project_members  │N─────1│   projects   │
│          │       │                  │       │              │
│ _id (PK) │       │ _id (PK)         │       │ _id (PK)     │
│ name     │       │ project_id (ref) │       │ name         │
│ email    │       │ user_id (ref)    │       │ description  │
│ password │       │ joined_at        │       │ created_by (ref)│
│ role     │       └──────────────────┘       │ created_at   │
│ created_at│                                 │ updated_at   │
│ updated_at│                                 └──────┬───────┘
└─────┬────┘                                         │
      │                                              │1
      │1                                             │
      │          ┌──────────────────┐                │
      └─────────N│     tasks        │N───────────────┘
                 │                  │
                 │ _id (PK)         │
                 │ project_id (ref) │
                 │ title            │
                 │ description      │
                 │ status           │
                 │ priority         │
                 │ assigned_to (ref)│
                 │ deadline         │
                 │ created_by (ref) │
                 │ created_at       │
                 │ updated_at       │
                 └──────────────────┘
```

---

## 6. Pembagian Kerja

| Anggota | Peran Utama | Tanggung Jawab Kode |
|---|---|---|
| **Raihan Zulfa Kamal** (434241096) | Git Master & Core Architecture | Inisialisasi repo, `go.mod`, struktur folder, `entity/`, `config/database.go`, `middleware/auth.go`, `Dockerfile`, `docker-compose.yml`, review PR, merge ke `main` |
| **Ody Dzakwan Berwin** (434241100) | Repository Layer & System Analyst | Seluruh file `repository/` (interface + implementasi query MongoDB) |
| **Muhammad Fadhil Ilyas** (434241099) | Usecase Layer & Lead Documentation | Seluruh file `service/`, `dto/`, wiring `main.go`, kompilasi laporan akhir |

---

## 7. Struktur Folder (Clean Architecture)

Sesuai kesepakatan modul, tidak ada layer `controller` atau `handler`. Wiring rute HTTP dilakukan langsung di `main.go`.

```
gotask-api/
├── entity/                         # Domain entity + enum
│   ├── user.go                     # User, Role
│   ├── project.go                  # Project, ProjectMember
│   └── task.go                     # Task, TaskStatus, TaskPriority
├── repository/                     # Interface + implementasi MongoDB
│   ├── user_repository.go
│   ├── project_repository.go
│   ├── project_member_repository.go
│   └── task_repository.go
├── service/                        # Logika bisnis (usecase)
│   ├── user_service.go
│   ├── project_service.go
│   └── task_service.go
├── dto/                            # Request & response payload
│   ├── user_dto.go
│   ├── project_dto.go
│   └── task_dto.go
├── config/
│   └── database.go                 # Koneksi MongoDB
├── middleware/
│   └── auth.go                     # Verifikasi JWT + otorisasi role
├── docs/
│   └── rancangan_sistem.md
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
├── main.go                         # Entry point, DI, routing
└── README.md
```

---

## 8. Daftar Endpoint REST API

| Method | Endpoint | Akses |
|---|---|---|
| POST | `/api/register` | Public |
| POST | `/api/login` | Public |
| GET | `/api/health` | Public |
| GET | `/api/users` | Admin |
| GET | `/api/users/:id` | Authenticated |
| PUT | `/api/users/:id` | Self / Admin |
| DELETE | `/api/users/:id` | Admin |
| POST | `/api/projects` | Admin, Manager |
| GET | `/api/projects` | Authenticated |
| GET | `/api/projects/:id` | Authenticated |
| PUT | `/api/projects/:id` | Owner, Admin |
| DELETE | `/api/projects/:id` | Owner, Admin |
| POST | `/api/projects/:id/members` | Owner, Admin |
| GET | `/api/projects/:id/members` | Authenticated |
| DELETE | `/api/projects/:id/members/:user_id` | Owner, Admin |
| POST | `/api/projects/:id/tasks` | Owner, Admin |
| GET | `/api/projects/:id/tasks` | Authenticated |
| GET | `/api/tasks/me` | Authenticated |
| GET | `/api/tasks/:id` | Authenticated |
| PUT | `/api/tasks/:id` | Creator, Admin |
| PATCH | `/api/tasks/:id/status` | Assignee, Creator, Admin |
| DELETE | `/api/tasks/:id` | Creator, Admin |
| GET | `/api/dashboard` | Authenticated |
