# Task: Update Backend Plan to Match New Frontend Architecture

## Overview

The frontend plan has been simplified significantly - removing all AI/ML components, web app, and Google Drive sync. The backend now needs to be a full RESTful API service instead of a file-based sync system.

## Key Architecture Changes

### OLD Architecture (Being Removed):
- File-based sync via Google Drive (rclone/rsync)
- Cron job pulling files from Google Drive
- On-device AI classification using Pixel 9 TPU
- Web app + Android app
- Confidence scoring for AI classification
- Backend as "lightweight storage server"

### NEW Architecture (Target):
- Direct HTTP REST API communication
- Real-time sync when online (no batch file sync)
- Simple pattern matching (user-defined merchant → category mappings)
- Android app only (no web app)
- Manual categorization with learning system
- Backend as full API service with PostgreSQL

## Todo List

### Phase 1: Data Structure Updates
- [ ] Add `MerchantPattern` data structure (merchant → category mappings)
- [ ] Remove `confidence` field from `Expense` structure
- [ ] Remove `confidence` field from `MerchantInfo` structure
- [ ] Remove `classified` field from `Transaction` structure (no AI classification)
- [ ] Add `userId` field to relevant structures for multi-user support
- [ ] Update `SyncStatus` to support incremental sync instead of file-based sync

### Phase 2: Remove Deprecated Endpoints
- [ ] Remove `POST /api/sync/pull` endpoint (Google Drive pull)
- [ ] Remove `POST /api/classify` endpoint (fallback AI classification)
- [ ] Remove all Google Drive references from documentation

### Phase 3: Add New Authentication Endpoints
- [ ] Add `POST /api/auth/register` - User registration
- [ ] Add `POST /api/auth/login` - User login
- [ ] Update device registration to link to user accounts
- [ ] Add JWT authentication documentation

### Phase 4: Add MerchantPattern Endpoints
- [ ] Add `GET /api/merchant-patterns` - Get all user's patterns
- [ ] Add `POST /api/merchant-patterns` - Create new pattern (e.g., "Always categorize Amazon as Shopping")
- [ ] Add `PUT /api/merchant-patterns/:id` - Update existing pattern
- [ ] Add `DELETE /api/merchant-patterns/:id` - Delete pattern

### Phase 5: Update Sync Endpoints
- [ ] Update `POST /api/transactions/batch` for incremental sync
- [ ] Add `POST /api/expenses/batch` for batch expense sync
- [ ] Add `POST /api/sync/incremental` for delta sync support
- [ ] Add conflict resolution strategy documentation (last-write-wins with timestamp)
- [ ] Update `POST /api/sync/status` to track real-time sync instead of file sync

### Phase 6: Update Data Flow Documentation
- [ ] Remove "Sync to Cloud (rsync to Google Drive)" section
- [ ] Remove "Backend Processing (pull from Google Drive)" section
- [ ] Remove "On-Device Classification (Pixel 9 TPU)" references
- [ ] Add "Pattern Matching" section (client-side)
- [ ] Add "Real-time API Sync" section
- [ ] Update manual entry flow to sync immediately via API
- [ ] Add conflict resolution flow

### Phase 7: Update Overview and Architecture Description
- [ ] Remove "lightweight storage server" description
- [ ] Add "RESTful API service" description
- [ ] Remove web app references
- [ ] Remove AI/ML references
- [ ] Remove Google Drive sync references
- [ ] Update to reflect Android-only client
- [ ] Add multi-user support description

### Phase 8: Add New Required Endpoints
- [ ] Add `GET /api/categories` (already exists, verify)
- [ ] Add `POST /api/categories` - Create custom category
- [ ] Add `PUT /api/categories/:id` - Update category
- [ ] Add `DELETE /api/categories/:id` - Delete category
- [ ] Add batch endpoints for offline sync support

## Execution Plan

### Step 1: Review Current Backend Plan
- Read through BACKEND_API.md thoroughly to understand all current endpoints and data structures

### Step 2: Create Updated Data Structures Section
- Add MerchantPattern structure
- Remove confidence scoring from all structures
- Add userId fields for multi-user support
- Update SyncStatus for real-time sync

### Step 3: Update/Remove Endpoints
- Remove deprecated endpoints (Google Drive sync, AI classification)
- Add authentication endpoints
- Add MerchantPattern CRUD endpoints
- Add batch sync endpoints
- Update existing endpoints documentation

### Step 4: Rewrite Data Flow Section
- Document new flow: Notification → Parse → Pattern Match → Local DB → HTTP API → PostgreSQL
- Add conflict resolution strategy
- Add offline sync handling
- Remove all Google Drive and AI references

### Step 5: Update Overview and Architecture
- Change architecture description from "lightweight storage" to "full REST API"
- Remove web app mentions
- Add user management and authentication details
- Document real-time sync vs offline batch sync

### Step 6: Final Review
- Ensure all AI/ML references removed
- Ensure all Google Drive references removed
- Ensure all web app references removed
- Verify all new required endpoints are documented
- Check that authentication is properly documented

## Notes

- The backend is now a **full API service**, not just a sync target
- No more Google Drive integration needed
- No more AI/ML endpoints or confidence scores
- Android app is the only client (no web app)
- Pattern matching happens on client-side using user-defined rules
- Real-time sync when online, batch sync when offline
- Need proper conflict resolution for offline edits

## Timeline

This is a documentation task, should take approximately 1-2 hours to complete thoroughly.

## Acceptance Criteria

- [x] All AI/ML components removed from documentation
- [x] All Google Drive sync removed from documentation
- [x] All web app references removed
- [x] New authentication system documented
- [x] MerchantPattern system documented
- [x] Real-time sync + conflict resolution documented
- [x] Backend positioned as full RESTful API service
- [x] All endpoints align with new frontend architecture

## COMPLETED

All tasks have been successfully completed. The backend API documentation has been fully updated to match the new frontend architecture.
