# Entaingo - Golang Application for Processing Incoming Requests

## Overview

Entaingo is a Golang application designed to process incoming POST requests from third-party providers. It manages user account balances based on the transaction status (win/lost) and supports Docker for easy deployment and testing.

NB- `for tests purposes I have include .env and app.env files fine for tests but very wrong for production`

## Technologies

- **Golang**
- **PostgreSQL**


## Requirements

1. **Processing and Saving Incoming Requests:**
   - The application exposes an HTTP endpoint to receive incoming POST requests.
   - Each request must include a JSON body with the following structure:
     ```json
      {
         "state":"win",
         "amount": 30.5,
         "transactionId": "txadv456"
      }
      ```
   - The `Source-Type` header can be one of three types: `game`, `server`, or `payment`.
   - Win requests increase the user balance, while lost requests decrease it.
   - Each transaction (identified by `transactionId`) is processed only once.
   - The account balance cannot be negative.

2. **Post-processing:**
   - Every N minutes, the application will cancel the 10 latest odd records and correct the user balance accordingly.
   - Canceled records must not be processed again.

3. **Docker Support:**
   - The application is prepared to run via Docker containers.

## Getting Started

### Prerequisites

- Go (1.20 or later)
- Docker
- Docker Compose
- PostgreSQL

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/myrachanto/entaingo.git
   cd entaingo
   ```

## Running the Application
### To Run the Application Locally


## Using Docker Compose
To build and run the application using Docker Compose, 
With this command you get to spin two docker containers the application and the postgress database. 

run:

```bash
make dockerCompose
``` 

and the application is live!

- In this repository, I've included a Postman collection. Once imported into Postman, it allows seamless access for testing the application.

## API Endpoints

```bash
POST localhost:4000/transaction/
```
## Request Headers:

Source-Type: client (game, server, payment)
Content-Type: application/json
Request Body:

```json
{
    "state":"win",
    "amount": 30.5,
    "transactionId": "txadv456"
}
```
- import the postman collection and run the post request while the application is running

## Responses:

- 200 OK: Successfully processed the request.
```json
{
    "data": {
        "user": {
            "id": 1,
            "balance": 70.5
        },
        "transaction": [
            {
                "id": 8,
                "transaction_id": "txadv456",
                "amount": 30.5,
                "state": "win",
                "source_type": "server",
                "user_id": 1,
                "processed_at": "2024-10-22T02:08:08.150753433Z",
                "canceled": false
            }
        ]
    }
}
```

- 400 Bad Request: Invalid input 
- 500 Internal Server Error

## other comands include


## Running Tests
To Run the Tests
Execute the following command:

```bash
make test
```

For Test Coverage
Run:

```bash
make testCover
```

## Generating Swagger Documentation
To generate Swagger documentation, use:

```bash
make swagger

```
```bash
http://localhost:4000/swagger/index.html
```

## Post-Processing
The application automatically cancels the 10 latest odd records every N minutes and adjusts the user balances. `in goroutine`



