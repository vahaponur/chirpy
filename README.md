# Chirpy

It is a little API project for purpose of learning web servers in golang.

## Installation



```bash
gh repo clone vahaponur/chirpy
```

## Usage
First make a .env file and put this or anything you like
```
JWT_SECRET=yourjwtsecret
POLKA_KEY=fakepolkakey
```
After that build project in debug mode, I used a fake json file for implementing database logic. If you don't build it with ```debug``` flag it will not delete fake db when you shut down the server
```bash
go build -o ./bin/chirpy && ./bin/chirpy --debug

```
# API Documentation

Welcome to the API documentation for the Chirpy application. This document provides information on how to use the API to perform various operations related to chirps and user management.

## Table of Contents

- [Chirp Operations](#chirp-operations)
    - [Create a Chirp](#create-a-chirp)
    - [Get Chirps](#get-chirps)
    - [Get Chirp by ID](#get-chirp-by-id)
    - [Delete Chirp by ID](#delete-chirp-by-id)
    - [Get Chirps by Author](#get-chirps-by-author)
    - [Get Chirps with Sorting](#get-chirps-with-sorting)
    - [Get Chirps by Author with Sorting](#get-chirps-by-author-with-sorting)
- [User Operations](#user-operations)
    - [Register User](#register-user)
    - [Login](#login)
    - [Update User Profile](#update-user-profile)
    - [Refresh Access Token](#refresh-access-token)
    - [Revoke Refresh Token](#revoke-refresh-token)
- [Polka Operations](#polka-operations)
    - [Upgrade User](#upgrade-user)

## Chirp Operations

### Create a Chirp

**Route:** `POST /chirps`

Create a new chirp.

**Request:**
- Method: `POST`
- Headers: Authentication token in the `Authorization` header
- Body: Chirp data in JSON format

**Response:**
- Status: `201 Created`
- Body: Created chirp details in JSON format

### Get Chirps

**Route:** `GET /chirps`

Get a list of chirps.

**Request:**
- Method: `GET`
- Headers: None
- Query Parameters: `sort` (Optional: `desc` or `asc`)

**Response:**
- Status: `200 OK`
- Body: List of chirps in JSON format

### Get Chirp by ID

**Route:** `GET /chirps/{id}`

Get details of a specific chirp by ID.

**Request:**
- Method: `GET`
- Headers: None
- Path Parameters: `id` (Chirp ID)

**Response:**
- Status: `200 OK`
- Body: Chirp details in JSON format

### Delete Chirp by ID

**Route:** `DELETE /chirps/{id}`

Delete a specific chirp by ID.

**Request:**
- Method: `DELETE`
- Headers: Authentication token in the `Authorization` header
- Path Parameters: `id` (Chirp ID)

**Response:**
- Status: `204 No Content`

### Get Chirps by Author

**Route:** `GET /chirps`

Get chirps by a specific author.

**Request:**
- Method: `GET`
- Headers: Authorization token in the `Authorization` header
- Query Parameters:
    - `author_id` (Author ID)
    - `sort` (Optional: `desc` or `asc`)

**Response:**
- Status: `200 OK`
- Body: List of chirps by the specified author in JSON format

### Get Chirps with Sorting

**Route:** `GET /chirps`

Get chirps with optional sorting.

**Request:**
- Method: `GET`
- Headers: None
- Query Parameters:
    - `sort` (Optional: `desc` or `asc`)

**Response:**
- Status: `200 OK`
- Body: List of chirps with optional sorting in JSON format

### Get Chirps by Author with Sorting

**Route:** `GET /chirps`

Get chirps by a specific author with optional sorting.

**Request:**
- Method: `GET`
- Headers: Authorization token in the `Authorization` header
- Query Parameters:
    - `author_id` (Author ID)
    - `sort` (Optional: `desc` or `asc`)

**Response:**
- Status: `200 OK`
- Body: List of chirps by the specified author with optional sorting in JSON format

## User Operations

### Register User

**Route:** `POST /users`

Register a new user.

**Request:**
- Method: `POST`
- Headers: None
- Body: User registration data in JSON format

**Response:**
- Status: `201 Created`
- Body: Created user details in JSON format

### Login

**Route:** `POST /login`

Log in a user and obtain authentication tokens.

**Request:**
- Method: `POST`
- Headers: None
- Body: User login credentials in JSON format

**Response:**
- Status: `200 OK`
- Body: User details along with access and refresh tokens in JSON format

### Update User Profile

**Route:** `PUT /users`

Update user profile information.

**Request:**
- Method: `PUT`
- Headers: Authentication token in the `Authorization` header
- Body: Updated user data in JSON format

**Response:**
- Status: `200 OK`
- Body: Updated user details in JSON format

### Refresh Access Token

**Route:** `POST /refresh`

Exchange a refresh token for a new access token.

**Request:**
- Method: `POST`
- Headers: Authentication token in the `Authorization` header
- Body: Refresh token in JSON format

**Response:**
- Status: `200 OK`
- Body: New access token in JSON format

### Revoke Refresh Token

**Route:** `POST /revoke`

Revoke a refresh token to invalidate it.

**Request:**
- Method: `POST`
- Headers: Authentication token in the `Authorization` header
- Body: Refresh token in JSON format

**Response:**
- Status: `204 No Content`

## Polka Operations

### Upgrade User

**Route:** `POST /polka/webhooks`

Handle a Polka event to upgrade a user.

**Request:**
- Method: `POST`
- Headers: Authentication token in the `Authorization` header
- Body: Polka event data in JSON format

**Response:**
- Status: `200 OK` (if event is not 'user.upgraded')
- Status: `200 OK` (if user is upgraded successfully)

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.
