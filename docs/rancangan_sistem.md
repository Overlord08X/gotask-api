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

### Anggota A — Git Master
- Setup project (init repo, struktur folder, config)
- Autentikasi: register & login
- Entity, Repository, Service untuk `User`
- Middleware JWT

### Anggota B
- Entity, Repository, Service untuk `Project`
- Manajemen Project Member (assign/remove member)

### Anggota C
- Entity, Repository, Service untuk `Task`
- Status tracking (update status task)
- Dashboard (ringkasan project & task per status)

---

## 7. Struktur Folder (Clean Architecture)

```
.
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── entity/                 # Domain entities
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── project_member.go
│   │   └── task.go
│   ├── repository/             # Interface & implementasi akses data
│   │   ├── user_repository.go
│   │   ├── project_repository.go
│   │   ├── project_member_repository.go
│   │   └── task_repository.go
│   ├── service/                # Business logic
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── project_service.go
│   │   ├── project_member_service.go
│   │   ├── task_service.go
│   │   └── dashboard_service.go
│   ├── dto/                    # Request & response struct
│   │   ├── auth_dto.go
│   │   ├── user_dto.go
│   │   ├── project_dto.go
│   │   ├── task_dto.go
│   │   └── dashboard_dto.go
│   ├── handler/                # HTTP handler (Gin)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── project_handler.go
│   │   ├── task_handler.go
│   │   └── dashboard_handler.go
│   └── middleware/             # JWT auth middleware
│       └── auth_middleware.go
├── config/
│   └── config.go               # Konfigurasi MongoDB, JWT secret, dll
├── docs/
│   └── rancangan_sistem.md
├── go.mod
└── go.sum
```
