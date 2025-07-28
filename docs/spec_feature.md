# Feature Specification

## 1. Overview

The Vera Identity Service provides authentication and authorization capabilities for the Vera ecosystem, supporting human users.

## 2. User Stories

### 2.1 User Authentication

**As a** user  
**I want to** authenticate using my Google account  
**So that** I can access Vera services securely  

**Acceptance Criteria:**

- User can initiate login via Google OAuth 2.0
- Only users who have been created in the system (with their email) can authenticate
- User receives two JWT tokens upon successful authentication:
  - Access token (short-lived, 15-60 minutes)
  - Refresh token (long-lived, up to 1 year)
- User can use refresh token to obtain new access tokens
- Access tokens are used for API calls
- Refresh tokens are used only for token renewal

### 2.2 User Management

**As a** system administrator  
**I want to** create user accounts for authorized individuals  
**So that** they can access Vera services  

**Acceptance Criteria:**

- Administrators can create user accounts by specifying email addresses
- User accounts are created with an email
- User creation timestamps are automatically recorded
- Failed authentication attempts for non-existent users are logged for security monitoring

## 3. Business Rules

### 3.1 User Access Control

- Only users who have been explicitly created in the system can authenticate
- User accounts are managed by system administrators
- Failed authentication attempts are logged for security monitoring
- User accounts can be soft-deleted (marked as deleted without removing data)

### 3.2 Token Management

- Access tokens are short-lived for security
- Refresh tokens allow long-term access without re-authentication
- Token usage is logged for audit purposes
- User login activity is tracked with timestamps and OAuth sub identifiers
