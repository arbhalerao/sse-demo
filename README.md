# sse-demo

A simple demo showcasing Server-Sent Events (SSE) with a Go backend and React frontend, enabling real-time one-way communication from server to client.

## Quick Start

```bash
# Start the application
docker-compose up --build

# Access the app at http://localhost:3000
```

## Endpoints

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Health Check: http://localhost:8080/health

## Testing

1. Open http://localhost:3000
2. Click "Send Ping" to test messaging
3. Watch real-time updates in the message list
4. Check connection status indicator

Built using React, Vite, and Go.
