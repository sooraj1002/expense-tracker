# Backend Plan & Progress - Automated Expense Tracking

## Task Overview
Update backend API to support automated expense tracking with notification reading, on-device AI classification, merchant detection, location fallback, and Google Drive sync.

## Current State Analysis

### Existing Backend Plan
✅ **BACKEND_API.md** - Basic expense CRUD API existed

### Requirements from CLAUDE.md
- Lightweight storage server (Golang)
- Backend just stores data, Android app does heavy lifting
- Support for auto-detected transactions from notifications
- Merchant information storage
- Location data for fallback
- Sync operations with Google Drive
- Transaction classification fallback API

## Implementation Plan

### Backend Updates Completed ✅
- [x] Add overview section explaining architecture
- [x] Add Transaction data structure (raw notification data)
- [x] Add MerchantInfo data structure
- [x] Add Location data structure
- [x] Add SyncStatus data structure
- [x] Update Expense model with auto-detection fields (source, merchantId, locationId, confidence, rawData, verified)
- [x] Add transaction endpoints (POST, batch POST, GET)
- [x] Add merchant endpoints (GET, POST, GET by ID with expenses)
- [x] Add location endpoints (POST, GET by ID with expenses)
- [x] Add sync endpoints (status update, status get, pull from Google Drive)
- [x] Add classification endpoint (fallback when on-device AI fails)
- [x] Add expense verification endpoint (PUT /api/expenses/:id/verify)
- [x] Document data flow (notification → parsing → classification → sync → backend)
- [x] Add authentication section (JWT with device registration)

## Updated API Endpoints Summary

### New Endpoints Added

#### Transactions
- `POST /api/transactions` - Create raw transaction from notification
- `POST /api/transactions/batch` - Batch upload during sync
- `GET /api/transactions` - Retrieve raw transactions with filters

#### Merchants
- `GET /api/merchants` - List detected merchants
- `POST /api/merchants` - Create/update merchant info
- `GET /api/merchants/:id/expenses` - Get merchant spending analysis

#### Locations
- `POST /api/locations` - Record location (fallback)
- `GET /api/locations/:id/expenses` - Get location-based expenses

#### Sync Operations
- `POST /api/sync/status` - Update sync status from device
- `GET /api/sync/status/:deviceId` - Get current sync status
- `POST /api/sync/pull` - Backend pulls from Google Drive (cron/manual)

#### Classification & Verification
- `POST /api/classify` - Classify transaction (fallback)
- `PUT /api/expenses/:id/verify` - Verify/correct auto-detected expense

#### Authentication
- `POST /api/auth/register` - Device registration with JWT

## Data Structures Added

### Transaction
Raw notification/SMS data before classification:
- id, rawText, timestamp, senderInfo, amount
- parsed, classified, expenseId (link to created expense)

### MerchantInfo
Detected merchant information:
- id, name, category, confidence
- aliases (name variations)
- commonCategoryId (most frequent category)

### Location
GPS fallback when merchant unclear:
- id, latitude, longitude, timestamp
- address (reverse geocoded)
- accuracy

### SyncStatus
Track sync state between device and backend:
- id, deviceId, lastSyncTime
- pendingCount, syncedCount
- status (syncing/completed/failed)
- errorMessage

## Data Flow Documented

### Automated Transaction Processing
1. **Notification Capture** (Android) → Extract transaction details → Local SQLite
2. **On-Device Classification** (Android) → Merchant detection → Category classification → Local storage
3. **Sync to Cloud** (Android → Google Drive) → Batch upload via rsync → Mark synced
4. **Backend Processing** (Google Drive → Backend) → Cron pulls data → Store in PostgreSQL
5. **API Access** (Frontend → Backend) → Query data → Display analytics

## Backend Technology Stack

### Core
- **Language:** Golang
- **Framework:** Gin (REST API)
- **Database:** PostgreSQL
- **Authentication:** JWT tokens

### Integrations
- **Google Drive API** - Pull synced data from mobile devices
- **Cron Jobs** - Periodic sync pull from Google Drive

### API Features
- RESTful endpoints
- Pagination support
- Filtering and search
- Batch operations
- Error handling

## Implementation Todos (After Approval)

### Phase 1: Database Schema
- [ ] Create PostgreSQL migrations
- [ ] Define tables: expenses, transactions, merchants, locations, sync_status, categories, accounts
- [ ] Add indexes for performance
- [ ] Setup foreign key relationships

### Phase 2: Core API Implementation
- [ ] Setup Golang project with Gin framework
- [ ] Implement authentication (JWT, device registration)
- [ ] Implement CRUD for all data models
- [ ] Add pagination and filtering

### Phase 3: Transaction Processing
- [ ] Implement transaction endpoints
- [ ] Implement merchant endpoints
- [ ] Implement location endpoints
- [ ] Add classification fallback logic

### Phase 4: Sync Engine
- [ ] Integrate Google Drive API
- [ ] Implement sync pull endpoint
- [ ] Create cron job for periodic pulls
- [ ] Handle conflict resolution
- [ ] Update sync status tracking

### Phase 5: Testing & Optimization
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Performance optimization
- [ ] Add logging and monitoring
- [ ] Security audit

### Phase 6: Deployment
- [ ] Setup CI/CD pipeline
- [ ] Deploy to cloud (AWS/GCP)
- [ ] Configure database backups
- [ ] Setup monitoring and alerts

## Key Design Decisions

### Lightweight Backend
- Backend is intentionally minimal
- Android app does heavy lifting (parsing, classification)
- Backend just stores and serves data
- Reduces server load and costs

### Sync Strategy
- Pull-based: Backend pulls from Google Drive
- Avoids direct device-to-server sync (better for battery)
- Google Drive acts as intermediary buffer
- Periodic cron jobs for pulling

### Data Integrity
- Transaction table preserves raw data
- Allows re-classification if AI improves
- Audit trail for all auto-detected expenses

### Scalability
- Designed for single-user or small teams
- Can scale horizontally if needed
- Database indexes for performance

## Security Considerations

### Authentication
- JWT tokens for API access
- Device-based registration
- Token refresh mechanism

### Data Privacy
- Backend only stores aggregated data
- No raw bank notification text stored permanently
- User can delete all data

### API Security
- HTTPS only
- Rate limiting
- Input validation
- SQL injection prevention

## Notes
- All AI classification happens on-device (Android)
- Backend provides fallback classification if needed
- Sync is designed for reliability (WiFi-only, batch operations)
- Backend ready for future enhancements (budgeting, analytics)

## Status
✅ **Planning Complete** - Backend API plan updated and documented
⏳ **Awaiting Approval** - Ready to begin implementation after review
