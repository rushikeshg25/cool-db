# Cooldb: Technical Development Plan

This document outlines the phased development plan for **Cooldb**, focusing on technical milestones, architectural improvements, and feature expansion.

---

## Phase 1: Foundation (Current Status: In Progress)
*Focus: Establish core communication and basic storage mechanics.*

- [x] **Core gRPC Server**: Basic implementation of the gRPC service layer using `cool-wire`.
- [x] **Unified CLI**: `cool server`, `cool shell`, and `cool exec` in one binary.
- [x] **Persistent File Management**: Atomic database snapshots with restart persistence.
- [x] **Query Pipeline**: CLI → gRPC → Core Engine → persistent storage.
- [x] **Basic Query Validation**: Parsing, type checking, and core constraints for the v0.1 SQL subset.

---

## Phase 2: Reliability & Persistence
*Focus: Transition from a volatile state to a durable database system.*

### 1. Write-Ahead Logging (WAL)
- Implement a sequential log to record all mutations before applying them to the main data files.
- Ensure recovery mechanisms are in place to replay the log after an unexpected crash.

### 2. Dot Commands & Admin Tools
- `.open <db_file>`: Switch between active database files.
- `.close`: Safely close the active database and sync buffers.
- `.dbinfo`: Provide metadata about the current database (version, size, record count).
- `.quit` / `.exit`: Implement graceful shutdown to ensure no data loss.

### 3. Error Handling Framework
- Extend the unified error system in `internal/database/errors.go` for cross-component reporting.
- Implement gRPC status codes for meaningful client-side feedback.

---

## Phase 3: Performance & Optimization
*Focus: Scale the system for efficiency and concurrency.*

### 1. Indexing (B-Tree implementation)
- Introduce indexing to allow O(log n) lookups instead of full table scans.
- Support for primary key indexes on `.db` files.

### 2. In-Memory Page Cache
- Implement a LRU (Least Recently Used) cache to keep hot data in memory.
- Reduce disk I/O for frequent read operations.

### 3. Concurrency Control
- Implement reader-writer locks to allow multiple concurrent reads.
- Ensure thread-safety for the storage engine.

---

## Phase 4: Ecosystem & Advanced Features
*Focus: Security, usability, and deployment readiness.*

### 1. Authentication Layer
- Simple token-based authentication for gRPC connections.
- Secure access control for the UI dashboard.

### 2. Automated Backups & Maintenance
- `.backup <destination>`: Online backup capability without stopping the server.
- Automatic log rotation and cleanup.

### 3. UI Dashboard v2
- Visual query builder.
- Performance monitoring charts (memory usage, query latency).
- Database schema visualization.

### 4. Deployment & DevOps
- Prepare `Dockerfile` for easy containerization.
- Implement CI/CD pipelines for automated testing and security scanning.
